Author: CkxNhật
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