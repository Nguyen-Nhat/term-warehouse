func (s *WSService) ReplaceSerialOfMaterial(ctx context.Context, request *ws.ReplaceSerialOfMaterialRequest) (*ws.ReplaceSerialOfMaterialResponse, error) {
	type skuSerialKey struct {
		sku    string // case lot
		serial string
	}
	var (
		logger                 = s.log.WithName("ReplaceSerialOfMaterial").WithValues("request", request)
		originRoundId          = request.GetRoundId()
		binId                  = request.GetBinId()
		mapOldSerial2NewSerial = make(map[skuSerialKey]string)  // data from item of request.Item
		mapSkuSerial2Quantity  = make(map[skuSerialKey]float64) // data from input item of RoundId
		disassembleInput       []*msApi.ManufactureSessionRoundItem
		disassembleOutput      []*msApi.ManufactureSessionRoundItem
		assembleInput          []*msApi.ManufactureSessionRoundItem
		assembleOutput         []*msApi.ManufactureSessionRoundItem
	)
	logger.Info("Start process")

	userInfo, err := helpers.GetUserInfo(ctx, helpers.EnableMultiSite())
	if err != nil {
		logger.Error(err, "GetUserInfo failed")
		return nil, err
	}

	rounds, err := s.manufactureClient.GetManufactureSessionRounds(ctx, &msApi.GetManufactureSessionRoundsRequest{
		SellerId: int32(userInfo.SellerId),
		RoundIds: slice.Make(request.RoundId),
		Offset:   EmptyString,
		Limit:    LimitOne,
	})
	if err != nil {
		logger.Error(err, "GetManufactureSessionRound failed")
		return nil, err
	}
	round := slice.GetFirst(rounds.GetData().GetItems())
	if round.GetRoundId() == EmptyString {
		err = status.Errorf(codes.NotFound, "Not found RoundId %v", request.GetRoundId())
		logger.Error(err, "Not found roundId")
		return nil, err
	}

	// check round status
	if round.GetStatus() != msApi.MANUFACTURE_SESSION_ROUND_STATUS_DONE {
		err = status.Errorf(codes.FailedPrecondition, "Status of RoundId %v is not done", request.GetRoundId())
		logger.Error(err, "status != done")
		return nil, err
	}

	// START:  validate data
	slice.ForEach(round.GetInputs(), func(item *msApi.ManufactureSessionRoundItem) {
		for _, serial := range item.GetSerials() {
			key := skuSerialKey{
				sku:    item.GetSku(),
				serial: serial,
			}
			mapSkuSerial2Quantity[key] = OneQuantity
		}
		for _, lot := range item.GetLots() {
			key := skuSerialKey{
				sku:    item.GetSku(),
				serial: lot.GetName(),
			}
			mapSkuSerial2Quantity[key] = lot.GetQuantity()
		}
	})
	isValidOldSerial := true
	slice.ForEach(request.GetItems(), func(item *ws.ReplaceSerialOfMaterialRequest_Item) {
		mapOldSerial2NewSerial[skuSerialKey{sku: item.GetSku(), serial: item.GetOldSerial()}] = item.GetNewSerial()
		if _, ok := mapSkuSerial2Quantity[skuSerialKey{sku: item.GetSku(), serial: item.GetOldSerial()}]; !ok {
			isValidOldSerial = false
		}
	})
	if !isValidOldSerial {
		err = status.Errorf(codes.FailedPrecondition, "Serial of Sku is not exist in RoundId %v", request.GetRoundId())
		logger.Error(err, "status != done")
		return nil, err
	}
	// END: validate data

	// START: prepare data
	roundCloneForDisassemble, ok := proto.Clone(round).(*msApi.GetManufactureSessionRoundsResponse_ManufactureSessionRound)
	if !ok {
		return nil, status.Error(codes.Internal, "internal error")
	}
	roundCloneForAssemble, ok := proto.Clone(round).(*msApi.GetManufactureSessionRoundsResponse_ManufactureSessionRound)
	if !ok {
		return nil, status.Error(codes.Internal, "internal error")
	}
	disassembleInput = roundCloneForDisassemble.GetOutputs()
	disassembleOutput = roundCloneForDisassemble.GetInputs()

	assembleInput = roundCloneForAssemble.GetInputs()
	assembleOutput = roundCloneForAssemble.GetOutputs()

	slice.ForEach(assembleInput, func(item *msApi.ManufactureSessionRoundItem) {
		for index, oldSerial := range item.Serials {
			if newSerial, ok := mapOldSerial2NewSerial[skuSerialKey{sku: item.GetSku(), serial: oldSerial}]; ok {
				item.Serials[index] = newSerial
			}
		}
		for index, oldLot := range item.Lots {
			if newLot, ok := mapOldSerial2NewSerial[skuSerialKey{sku: item.GetSku(), serial: oldLot.GetName()}]; ok {
				item.Lots[index].Name = newLot
			}
		}
	})
	// END: prepare data

	sessions, err := s.manufactureClient.GetManufactureSessions(ctx, &msApi.GetManufactureSessionsRequest{
		SellerId:   int32(userInfo.SellerId),
		SessionIds: slice.Make(round.GetSessionId()),
		Offset:     EmptyString,
		Limit:      LimitOne,
	})
	if err != nil {
		logger.Error(err, "GetManufactureSessions failed")
		return nil, err
	}
	session := slice.GetFirst(sessions.GetData().GetManufactureSessions())
	if session.GetSessionId() == EmptyString {
		err = status.Errorf(codes.NotFound, "Not found Session with RuondId %v", request.GetRoundId())
		logger.Error(err, "Not found sessionId")
		return nil, err
	}
	originRoundId = math.TernaryOp(stringz.IsEmpty(session.GetOriginRoundId()), originRoundId, session.GetOriginRoundId())

	createSessionResp, err := s.manufactureClient.CreateManufactureSession(ctx, &msApi.CreateManufactureSessionRequest{
		SellerId:      int32(userInfo.SellerId),
		OriginRoundId: originRoundId,
		Outputs: slice.Map(request.GetItems(), func(item *ws.ReplaceSerialOfMaterialRequest_Item) *msApi.CreateManufactureSessionRequest_Item {
			return &msApi.CreateManufactureSessionRequest_Item{
				Sku:           item.GetSku(),
				SerialLotName: item.GetNewSerial(),
				Quantity:      mapSkuSerial2Quantity[skuSerialKey{sku: item.GetSku(), serial: item.GetOldSerial()}],
			}
		}),
		Inputs: slice.Map(request.GetItems(), func(item *ws.ReplaceSerialOfMaterialRequest_Item) *msApi.CreateManufactureSessionRequest_Item {
			return &msApi.CreateManufactureSessionRequest_Item{
				Sku:           item.GetSku(),
				SerialLotName: item.GetOldSerial(),
				Quantity:      mapSkuSerial2Quantity[skuSerialKey{sku: item.GetSku(), serial: item.GetOldSerial()}],
			}
		}),
		SiteId:      session.GetSiteId(),
		SessionType: msApi.SESSION_TYPE_REPLACE_MATERIAL,
		CreatedBy:   userInfo.Sub,
		Note:        request.GetNote(),
		BinId:       binId,
	})
	if err != nil {
		logger.Error(err, "CreateManufactureSession failed")
		return nil, err
	}

	err = s.processMsRoundForReplaceSerialOfMaterial(ctx, processMsRoundForReplaceSerialOfMaterialRequest{
		sessionId: createSessionResp.GetData().GetSessionId(),
		binId:     binId,
		createBy:  userInfo.Sub,
		roundType: msApi.ROUND_TYPE_DISASSEMBLE,
		inputs:    disassembleInput,
		outputs:   disassembleOutput,
	})
	if err != nil {
		logger.Error(err, "processMsRoundForReplaceSerialOfMaterial roundType = disassemble failed")
		return nil, err
	}

	err = s.processMsRoundForReplaceSerialOfMaterial(ctx, processMsRoundForReplaceSerialOfMaterialRequest{
		sessionId: createSessionResp.GetData().GetSessionId(),
		binId:     binId,
		createBy:  userInfo.Sub,
		roundType: msApi.ROUND_TYPE_ASSEMBLE,
		inputs:    assembleInput,
		outputs:   assembleOutput,
	})
	if err != nil {
		logger.Error(err, "processMsRoundForReplaceSerialOfMaterial roundType = assemble failed")
		return nil, err
	}
	return &ws.ReplaceSerialOfMaterialResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
	}, nil
}

type processMsRoundForReplaceSerialOfMaterialRequest struct {
	sessionId string
	roundType msApi.RoundType
	createBy  string
	binId     int32
	inputs    []*msApi.ManufactureSessionRoundItem
	outputs   []*msApi.ManufactureSessionRoundItem
}

func (s *WSService) processMsRoundForReplaceSerialOfMaterial(ctx context.Context, params processMsRoundForReplaceSerialOfMaterialRequest) error {
	round, err := s.manufactureClient.UpsertManufactureSessionRound(ctx, &msApi.UpsertManufactureSessionRoundRequest{
		RoundId:   EmptyString,
		SessionId: params.sessionId,
		RoundType: params.roundType,
		Type:      msApi.MANUFACTURE_SESSION_ROUND_ACTION_TYPE_EXPORT,
		CreatedBy: params.createBy,
		BinId:     params.binId,
		Items:     params.inputs,
	})
	if err != nil {
		return err
	}

	_, err = s.manufactureClient.UpsertManufactureSessionRound(ctx, &msApi.UpsertManufactureSessionRoundRequest{
		RoundId:   round.GetData().GetRoundId(),
		SessionId: params.sessionId,
		RoundType: params.roundType,
		Type:      msApi.MANUFACTURE_SESSION_ROUND_ACTION_TYPE_IMPORT,
		CreatedBy: params.createBy,
		BinId:     params.binId,
		Items:     params.outputs,
	})
	return err
}
