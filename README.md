Author: CkxNhật
#  Overview 
   - **ERP**: Hệ thống OMNI
   - **COV**: Centralize Order View, service centralize các thông tin về đơn hàng cho mục đích hiển thị.
   - **FFR**: Fulfillment Router, service quản lý shipment (1 order có thể có nhiều shipment)
   - **WS** : Chịu trách nhiệm xử lý các nghiệp vụ của WMS
   - **TMDT, Marketplaces**: Là các sàn thương mại điện tử được tích hợp vào ERP

# Terms in Warehouse Service 

## **Yêu cầu xuất kho (ER - Export Request):**
1. **SO (Sales Order):**  
   - Đơn hàng bán.  
   - Thường đại diện cho yêu cầu xuất kho để giao sản phẩm đến khách hàng.
   - Prefix: FFR-

2. **ST (Stock Transfer):**  
   - Chuyển kho.  
   - Liên quan đến việc chuyển hàng từ kho này sang kho khác trong cùng hệ thống.
   - Prefix: ST-

3. **POR (Purchase Order Return):**  
   - Trả hàng mua.  
   - Dùng khi trả lại hàng đã mua cho nhà cung cấp.
   - Prefix: SC-OR-

---

## **Yêu cầu nhập kho (IR - Import Request):**
1. **PO (Purchase Order):**  
   - Đơn đặt hàng mua.  
   - Đại diện cho việc nhập hàng từ nhà cung cấp.
   - Prefix: SC-

2. **ST (Stock Transfer):**  
   - Chuyển kho.  
   - Dùng khi hàng hóa được nhập vào từ một kho khác trong hệ thống.
   - Prefix: ST-

3. **SOR (Sales Order Return):**  
   - Trả hàng bán.  
   - Khi khách hàng trả lại hàng hóa, nó được nhập lại vào kho.
   - Prefix: RM-
   - Từ RM -> FC-WMS -> WMS (ram events)
   - Contact Info: của sàn hoặc nhân viên, không có case nào thuộc khách hàng

## **Bin:**
1. **Type:**
   - B: type hàng bán
   - RO: type xe đẩy/ rỗ/ processing
   - PACK: type đóng gói 
   - BG: type bàn giao 
   
## Epic Note:   
## **OMNI-1386**: 
   - OMNI-1386 lúc trước design logic thế này đối với ST:
   - Xuất: Buộc phải xuất đúng loại hàng
   - Nhập: Buộc phải nhập vào loại hàng mặc định, loại hàng hiển thị lên phiếu nhập chỉ để user ref đến sau khi QC xong move vào

## **OMNI-1502**: 