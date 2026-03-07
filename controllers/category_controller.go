package controllers

import (
	"net/http"
	"strconv"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

// CategoryController maneja las operaciones de categorías
type CategoryController struct{}

// NewCategoryController crea una nueva instancia del controlador
func NewCategoryController() *CategoryController {
	return &CategoryController{}
}

// CategoryResponse estructura la respuesta de categoría
type CategoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

// GetCategories godoc
// @Summary      Listar categorías
// @Description  Obtiene todas las categorías activas disponibles en el sistema
// @Tags         categories
// @Produce      json
// @Success      200 {object} utils.Response{data=[]models.Category} "Lista de categorías"
// @Router       /categories [get]
// @Failure      401 {object} utils.Response "No autorizado"
func (cc *CategoryController) GetCategories(c *gin.Context) {
	var categories []models.Category

	query := config.DB.Where("is_active = ?", true).Order("sort_order ASC, name ASC")

	// Si es admin, puede ver todas
	if c.GetString("role") == "admin" {
		query = config.DB.Order("sort_order ASC, name ASC")
	}

	if err := query.Find(&categories).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener categorías", err.Error())
		return
	}

	// Transformar respuesta
	var response []CategoryResponse
	for _, cat := range categories {
		response = append(response, CategoryResponse{
			ID:          cat.ID,
			Name:        cat.Name,
			Slug:        cat.Slug,
			Description: cat.Description,
			Icon:        cat.Icon,
			Color:       cat.Color,
			IsActive:    cat.IsActive,
			SortOrder:   cat.SortOrder,
		})
	}

	utils.Success(c, http.StatusOK, "Lista de categorías", response)
}

// GetCategoryByID godoc
// @Summary      Obtener categoría por ID
// @Description  Obtiene los detalles de una categoría específica
// @Tags         categories
// @Produce      json
// @Param        id path int true "ID de la categoría"
// @Success      200 {object} utils.Response{data=models.Category} "Categoría encontrada"
// @Router       /categories/{id} [get]
// @Failure      400 {object} utils.Response "ID inválido"
// @Failure      404 {object} utils.Response "Categoría no encontrada"
func (cc *CategoryController) GetCategoryByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID inválido", nil)
		return
	}

	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Categoría no encontrada", nil)
		return
	}

	response := CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Icon:        category.Icon,
		Color:       category.Color,
		IsActive:    category.IsActive,
		SortOrder:   category.SortOrder,
	}

	utils.Success(c, http.StatusOK, "Categoría encontrada", response)
}

// CreateCategory godoc
// @Summary      Crear una categoría
// @Description  Crea una nueva categoría en el sistema (Solo Admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body CreateCategoryRequest true "Datos de la categoría"
// @Success      201 {object} utils.Response{data=models.Category} "Categoría creada exitosamente"
// @Failure      400 {object} utils.Response "Datos inválidos"
// @Failure      401 {object} utils.Response "No autorizado"
// @Failure      403 {object} utils.Response "Se requiere rol de administrador"
// @Router       /admin/categories [post]
// @Security     BearerAuth
// @example request -json {"name": "Carreras de Caballos", "description": "Torneos de pollas de caballos", "icon": "horse", "color": "#8B4513", "sort_order": 2}
// @example response -json {"success": true, "message": "Categoría creada exitosamente", "data": {"id": 1, "name": "Carreras de Caballos", "slug": "carreras-de-caballos", "is_active": true}}
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	var input CreateCategoryRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos: "+err.Error(), nil)
		return
	}

	// Generar slug automáticamente
	generatedSlug := slug.Make(input.Name)

	category := models.Category{
		Name:        input.Name,
		Slug:        generatedSlug,
		Description: input.Description,
		Icon:        input.Icon,
		Color:       input.Color,
		IsActive:    true,
		SortOrder:   input.SortOrder,
	}

	if err := config.DB.Create(&category).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al crear categoría", err.Error())
		return
	}

	response := CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Icon:        category.Icon,
		Color:       category.Color,
		IsActive:    category.IsActive,
		SortOrder:   category.SortOrder,
	}

	utils.Success(c, http.StatusCreated, "Categoría creada exitosamente", response)
}

// UpdateCategory godoc
// @Summary      Actualizar una categoría
// @Description  Actualiza los datos de una categoría existente (Solo Admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id path int true "ID de la categoría"
// @Param        request body UpdateCategoryRequest true "Datos a actualizar"
// @Success      200 {object} utils.Response{data=models.Category} "Categoría actualizada"
// @Failure      400 {object} utils.Response "Datos inválidos"
// @Failure      401 {object} utils.Response "No autorizado"
// @Failure      403 {object} utils.Response "Se requiere rol de administrador"
// @Failure      404 {object} utils.Response "Categoría no encontrada"
// @Router       /admin/categories/{id} [put]
// @Security     BearerAuth
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID inválido", nil)
		return
	}

	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Categoría no encontrada", nil)
		return
	}

	var input UpdateCategoryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", nil)
		return
	}

	// Actualizar campos
	updates := map[string]interface{}{}
	if input.Name != "" {
		updates["name"] = input.Name
		updates["slug"] = slug.Make(input.Name)
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Icon != "" {
		updates["icon"] = input.Icon
	}
	if input.Color != "" {
		updates["color"] = input.Color
	}
	if input.SortOrder > 0 {
		updates["sort_order"] = input.SortOrder
	}

	if err := config.DB.Model(&category).Updates(updates).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al actualizar categoría", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Categoría actualizada", category)
}

// DeleteCategory godoc
// @Summary      Eliminar una categoría
// @Description  Elimina una categoría del sistema (Solo Admin)
// @Tags         admin
// @Produce      json
// @Param        id path int true "ID de la categoría"
// @Success      200 {object} utils.Response "Categoría eliminada"
// @Failure      401 {object} utils.Response "No autorizado"
// @Failure      403 {object} utils.Response "Se requiere rol de administrador"
// @Failure      404 {object} utils.Response "Categoría no encontrada"
// @Router       /admin/categories/{id} [delete]
// @Security     BearerAuth
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID inválido", nil)
		return
	}

	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Categoría no encontrada", nil)
		return
	}

	if err := config.DB.Delete(&category).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al eliminar categoría", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Categoría eliminada", nil)
}

// ToggleCategoryStatus godoc
// @Summary      Cambiar estado de categoría
// @Description  Activa o desactiva una categoría (Solo Admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id path int true "ID de la categoría"
// @Param        request body ToggleCategoryStatusRequest true "Estado"
// @Success      200 {object} utils.Response{data=models.Category} "Estado actualizado"
// @Failure      400 {object} utils.Response "Datos inválidos"
// @Failure      401 {object} utils.Response "No autorizado"
// @Failure      403 {object} utils.Response "Se requiere rol de administrador"
// @Failure      404 {object} utils.Response "Categoría no encontrada"
// @Router       /admin/categories/{id}/status [patch]
// @Security     BearerAuth
func (cc *CategoryController) ToggleCategoryStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID inválido", nil)
		return
	}

	var input ToggleCategoryStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", nil)
		return
	}

	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Categoría no encontrada", nil)
		return
	}

	if err := config.DB.Model(&category).Update("is_active", input.IsActive).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al actualizar estado", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Estado actualizado", category)
}

// DTOs para las requests
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	SortOrder   int    `json:"sort_order"`
}

type ToggleCategoryStatusRequest struct {
	IsActive bool `json:"is_active"`
}
