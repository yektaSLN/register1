package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"login/dto"
	"login/models"
	"login/service"
	"login/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) Create(c *gin.Context) {

	userID, ok := getUserID(c)

	if !ok {
		h.handleError(c, utils.ErrInvalidToken)
		return
	}

	var request dto.CreateProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}

	product := &models.Product{
		Name:  request.Name,
		Price: request.Price,
	}

	if err := h.productService.Create(userID, product); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"product": product,
	})
}

func (h *ProductHandler) GetAll(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		h.handleError(c, utils.ErrInvalidToken)
		return
	}
	products, err := h.productService.GetAll(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"products":       products,
		"excel_download": "/api/products/export",
	})
}

func (h *ProductHandler) Export(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		h.handleError(c, utils.ErrInvalidToken)
		return
	}

	products, err := h.productService.GetAll(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	file := excelize.NewFile()
	defer file.Close()

	sheetName := "Products"
	file.SetSheetName("Sheet1", sheetName)

	headers := []string{
		"ID",
		"Name",
		"Price",
		"Created At",
		"Updated At",
	}

	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			h.handleError(c, err)
			return
		}

		if err := file.SetCellValue(sheetName, cell, header); err != nil {
			h.handleError(c, err)
			return
		}
	}

	for row, product := range products {
		values := []interface{}{
			row + 1,
			product.Name,
			product.Price,
			product.CreatedAt.Format("2006-01-02 15:04:05"),
			product.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		for col, value := range values {
			cell, err := excelize.CoordinatesToCellName(col+1, row+2)
			if err != nil {
				h.handleError(c, err)
				return
			}

			if err := file.SetCellValue(sheetName, cell, value); err != nil {
				h.handleError(c, err)
				return
			}
		}
	}

	file.SetColWidth(sheetName, "A", "A", 10)
	file.SetColWidth(sheetName, "B", "B", 30)
	file.SetColWidth(sheetName, "C", "C", 20)
	file.SetColWidth(sheetName, "D", "E", 25)

	filename := fmt.Sprintf("products_user_%d.xlsx", userID)

	c.Header(
		"Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	)

	c.Header(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, filename),
	)

	c.Header("Content-Transfer-Encoding", "binary")

	if err := file.Write(c.Writer); err != nil {
		return
	}
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		h.handleError(c, utils.ErrInvalidToken)
		return
	}
	productID, err := getProductID(c)
	if err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}
	product, err := h.productService.GetByID(
		userID,
		productID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"product": product,
	})
}

func (h *ProductHandler) Update(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		h.handleError(c, utils.ErrInvalidToken)
		return
	}
	productID, err := getProductID(c)
	if err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}
	var request dto.UpdateProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}
	product := &models.Product{
		Name:  request.Name,
		Price: request.Price,
	}
	if err := h.productService.Update(
		userID,
		productID,
		product,
	); err != nil {
		h.handleError(c, err)
		return
	}

	updatedProduct, err := h.productService.GetByID(
		userID,
		productID,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"product": updatedProduct,
	})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		h.handleError(c, utils.ErrInvalidToken)
		return
	}
	productID, err := getProductID(c)
	if err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}
	if err := h.productService.Delete(
		userID,
		productID,
	); err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "product deleted successfully",
	})
}

func getUserID(c *gin.Context) (uint, bool) {

	userIDValue, exists := c.Get("user_id")

	if !exists {
		return 0, false
	}

	userID, ok := userIDValue.(uint)

	return userID, ok
}

func getProductID(c *gin.Context) (uint, error) {

	id := c.Param("id")

	productID, err := strconv.ParseUint(id, 10, 64)

	if err != nil {
		return 0, err
	}

	if productID == 0 {
		return 0, errors.New("invalid product id")
	}

	return uint(productID), nil
}

func (h *ProductHandler) handleError(c *gin.Context, err error) {

	var appErr *utils.AppError

	if errors.As(err, &appErr) {
		c.JSON(
			appErr.StatusCode,
			utils.NewErrorResponse(appErr),
		)

		return
	}

	c.JSON(
		http.StatusInternalServerError,
		utils.NewErrorResponse(utils.ErrInternalServer),
	)
}
