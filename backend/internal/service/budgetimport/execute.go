package budgetimport

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type previewSnapshot struct {
	previewID string
	fileName  string
	fileData  []byte
	layoutKey string
	options   dto.PreviewBudgetImportOptions
	createdAt time.Time
}

type catalogRuntime struct {
	statuses     map[string]int64
	priorities   map[string]int64
	installers   map[string]int64
	projects     map[string]int64
	projectTypes map[string]int64
	salespeople  map[string]int64
	contacts     map[string]int64
	lossReasons  map[string]int64
}

func (s *service) storePreview(snapshot previewSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot.createdAt = time.Now()
	s.previews[snapshot.previewID] = snapshot
}

func (s *service) takePreview(previewID string) (previewSnapshot, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, exists := s.previews[previewID]
	if exists {
		delete(s.previews, previewID)
	}

	return snapshot, exists
}

func (s *service) ExecuteImport(ctx context.Context, req *dto.ExecuteBudgetImportRequest) (*dto.ExecuteBudgetImportResponse, error) {
	if req == nil || strings.TrimSpace(req.PreviewID) == "" {
		return nil, apperror.BadRequest("preview_id e obrigatorio")
	}

	snapshot, exists := s.takePreview(strings.TrimSpace(req.PreviewID))
	if !exists {
		return nil, apperror.NotFound("Preview nao encontrado")
	}

	layout, exists := findImportLayoutByKey(snapshot.layoutKey)
	if !exists {
		return nil, apperror.BadRequest("Layout de importacao nao encontrado para o preview")
	}
	importLogger := logger.FromContext(ctx).With(
		slog.String("import_action", "execute"),
		slog.String("preview_id", snapshot.previewID),
		slog.String("source_layout", layout.Key()),
		slog.String("source_company", layout.SourceCompany()),
		slog.String("file_name", snapshot.fileName),
	)
	importLogger.InfoContext(ctx, "execucao da importacao iniciada")

	workbook, err := loadWorkbookForLayout(snapshot.fileData, layout)
	if err != nil {
		return nil, err
	}

	catalogs, err := s.loadCatalogRuntime(ctx)
	if err != nil {
		return nil, err
	}

	startedAt := time.Now().UTC()
	summary := dto.ExecuteBudgetImportSummary{}
	status := "completed"
	resultMessage := "Importacao concluida com sucesso."
	header := workbook.rows[layout.HeaderRowNumber()]
	importBatchID, err := s.createImportBatch(ctx, snapshot, layout, startedAt)
	if err != nil {
		return nil, err
	}

	for rowNumber := layout.HeaderRowNumber() + 1; rowNumber <= workbook.maxRow; rowNumber++ {
		rowValues := workbook.rows[rowNumber]
		if layout.IsRowEmpty(rowValues) {
			continue
		}

		summary.RowsProcessed++

		row, rowErr := layout.ParseNormalizedRow(rowNumber, rowValues)
		if rowErr != nil {
			summary.RowsFailed++
			status = "completed_with_errors"
			if err := s.recordImportRow(ctx, importBatchID, header, rowValues, row, "error", "error", []string{rowErr.Error()}, 0); err != nil {
				return nil, err
			}
			continue
		}

		rowResult, importErr := s.importBudgetRow(ctx, catalogs, snapshot.options, layout, importBatchID, row)
		if importErr != nil {
			summary.RowsFailed++
			status = "completed_with_errors"
			if err := s.recordImportRow(ctx, importBatchID, header, rowValues, row, "error", "error", append(row.warnings, importErr.Error()), 0); err != nil {
				return nil, err
			}
			continue
		}

		summary.CatalogsCreated += rowResult.catalogsCreated
		rowStatus := "success"
		rowMessages := append([]string{}, row.warnings...)
		rowMessages = append(rowMessages, rowResult.messages...)
		if len(rowMessages) > 0 {
			rowStatus = "warning"
		}
		switch rowResult.action {
		case "create":
			summary.BudgetsCreated++
		case "update":
			summary.BudgetsUpdated++
		case "ignore":
			summary.BudgetsIgnored++
		}
		if err := s.recordImportRow(ctx, importBatchID, header, rowValues, row, rowStatus, rowResult.action, rowMessages, rowResult.budgetID); err != nil {
			return nil, err
		}
	}

	finishedAt := time.Now().UTC()
	if status == "completed_with_errors" {
		resultMessage = "Importacao concluida com inconsistencias."
	}
	if err := s.finishImportBatch(ctx, importBatchID, status, summary, resultMessage, finishedAt); err != nil {
		return nil, err
	}

	importID := fmt.Sprintf("import_%d", finishedAt.UnixNano())
	if importBatchID > 0 {
		importID = strconv.FormatInt(importBatchID, 10)
	}
	importLogger.InfoContext(
		ctx,
		"execucao da importacao concluida",
		slog.String("status", status),
		slog.String("import_id", importID),
		slog.Int("rows_processed", summary.RowsProcessed),
		slog.Int("budgets_created", summary.BudgetsCreated),
		slog.Int("budgets_updated", summary.BudgetsUpdated),
		slog.Int("budgets_ignored", summary.BudgetsIgnored),
		slog.Int("rows_failed", summary.RowsFailed),
		slog.Int("catalogs_created", summary.CatalogsCreated),
	)
	return &dto.ExecuteBudgetImportResponse{
		ImportID:   importID,
		PreviewID:  snapshot.previewID,
		Status:     status,
		StartedAt:  startedAt.Format(time.RFC3339),
		FinishedAt: finishedAt.Format(time.RFC3339),
		Summary:    summary,
		Result: dto.ExecuteBudgetImportResult{
			Message: resultMessage,
		},
	}, nil
}

type importRowResult struct {
	action          string
	catalogsCreated int
	budgetID        int64
	messages        []string
}

func (s *service) importBudgetRow(
	ctx context.Context,
	catalogs *catalogRuntime,
	options dto.PreviewBudgetImportOptions,
	layout importLayout,
	importBatchID int64,
	row normalizedBudgetImportRow,
) (importRowResult, error) {
	statusID, created, err := s.ensureBudgetStatusID(ctx, catalogs, row.statusName, options)
	if err != nil {
		return importRowResult{}, err
	}
	catalogsCreated := created

	priorityID, created, err := s.ensurePriorityID(ctx, catalogs, row.priorityName, options)
	if err != nil {
		return importRowResult{}, err
	}
	catalogsCreated += created

	projectTypeID := int64(0)
	projectID := int64(0)
	shouldAssociateProject := !shouldSkipProjectAssociation(row.projectName)

	if shouldAssociateProject {
		projectTypeID, created, err = s.ensureProjectTypeID(ctx, catalogs, row.projectTypeName, options)
		if err != nil {
			return importRowResult{}, err
		}
		catalogsCreated += created
	}

	installerID, created, err := s.ensureInstallerID(ctx, catalogs, row.installerName, options)
	if err != nil {
		return importRowResult{}, err
	}
	catalogsCreated += created

	if shouldAssociateProject {
		projectID, created, err = s.ensureProjectID(ctx, catalogs, row.projectName, projectTypeID, options)
		if err != nil {
			return importRowResult{}, err
		}
		catalogsCreated += created
	}

	salespersonID, created, err := s.ensureSalespersonID(ctx, catalogs, row.salespersonName, options)
	if err != nil {
		return importRowResult{}, err
	}
	catalogsCreated += created

	contactID, created, err := s.ensureContactID(ctx, catalogs, installerID, row.contactName, options)
	if err != nil {
		return importRowResult{}, err
	}
	catalogsCreated += created

	lossReasonID, created, err := s.ensureLossReasonID(ctx, catalogs, row.lossReasonName, options)
	if err != nil {
		return importRowResult{}, err
	}
	catalogsCreated += created

	existingBudget, err := s.budgetRepo.GetBySourceAndNumberAndYear(
		ctx,
		layout.SourceCompany(),
		row.budgetNumber,
		row.yearBudget,
	)
	if err != nil {
		return importRowResult{}, apperror.Internal("failed to check existing budget", err)
	}

	budgetModel := &model.BudgetModel{
		BudgetNumber:         row.budgetNumber,
		YearBudget:           row.yearBudget,
		Revision:             row.revision,
		SentAt:               row.sentAt,
		GrossValue:           row.grossValue,
		CommissionValue:      row.commissionValue,
		AreaM2:               row.areaM2,
		StatusID:             statusID,
		PriorityID:           validNullInt64(priorityID),
		InstallerID:          validNullInt64(installerID),
		ProjectID:            validNullInt64(projectID),
		SalespersonID:        validNullInt64(salespersonID),
		ContactID:            validNullInt64(contactID),
		LossReasonID:         validNullInt64(lossReasonID),
		CompetitorName:       row.competitorName,
		CompetitorPrice:      validNullFloat64(row.competitorPrice),
		DesignerName:         row.designerName,
		SpecificationDetails: row.specification,
		CurrentFollowUp:      row.currentFollowUp,
		SourceCompany:        layout.SourceCompany(),
		SourceLayout:         layout.Key(),
		ImportBatchID:        validNullInt64(importBatchID),
		UpdatedAt:            time.Now(),
	}

	if existingBudget != nil {
		if options.DuplicateStrategy == duplicateStrategyIgnore {
			return importRowResult{
				action:          "ignore",
				catalogsCreated: catalogsCreated,
				budgetID:        existingBudget.ID,
				messages:        []string{"Orcamento ignorado por duplicidade."},
			}, nil
		}

		budgetModel.ID = existingBudget.ID
		if err := s.budgetRepo.Update(ctx, budgetModel); err != nil {
			return importRowResult{}, apperror.Internal("failed to update budget", err)
		}

		return importRowResult{
			action:          "update",
			catalogsCreated: catalogsCreated,
			budgetID:        existingBudget.ID,
		}, nil
	}

	budgetModel.CreatedAt = budgetModel.UpdatedAt
	createdBudgetID, err := s.budgetRepo.Create(ctx, budgetModel)
	if err != nil {
		return importRowResult{}, apperror.Internal("failed to create budget", err)
	}

	return importRowResult{
		action:          "create",
		catalogsCreated: catalogsCreated,
		budgetID:        createdBudgetID,
	}, nil
}

func (s *service) createImportBatch(
	ctx context.Context,
	snapshot previewSnapshot,
	layout importLayout,
	startedAt time.Time,
) (int64, error) {
	userID := actorUserIDFromContext(ctx)
	batch := &model.BudgetImportBatchModel{
		PreviewID:        snapshot.previewID,
		FileName:         snapshot.fileName,
		SourceCompany:    layout.SourceCompany(),
		SourceLayout:     layout.Key(),
		Status:           "processing",
		ExecutedByUserID: validNullInt64(userID),
		StartedAt:        startedAt,
		ResultMessage:    "",
		CreatedAt:        startedAt,
		UpdatedAt:        startedAt,
	}

	batchID, err := s.auditRepo.CreateBatch(ctx, batch)
	if err != nil {
		return 0, apperror.Internal("failed to create import batch", err)
	}

	return batchID, nil
}

func (s *service) finishImportBatch(
	ctx context.Context,
	importBatchID int64,
	status string,
	summary dto.ExecuteBudgetImportSummary,
	resultMessage string,
	finishedAt time.Time,
) error {
	if importBatchID <= 0 {
		return nil
	}

	return s.auditRepo.UpdateBatch(ctx, &model.BudgetImportBatchModel{
		ID:              importBatchID,
		Status:          status,
		FinishedAt:      validNullTime(finishedAt),
		RowsProcessed:   summary.RowsProcessed,
		BudgetsCreated:  summary.BudgetsCreated,
		BudgetsUpdated:  summary.BudgetsUpdated,
		BudgetsIgnored:  summary.BudgetsIgnored,
		RowsFailed:      summary.RowsFailed,
		CatalogsCreated: summary.CatalogsCreated,
		ResultMessage:   resultMessage,
		UpdatedAt:       finishedAt,
	})
}

func (s *service) recordImportRow(
	ctx context.Context,
	importBatchID int64,
	header []string,
	rowValues []string,
	row normalizedBudgetImportRow,
	status string,
	action string,
	messages []string,
	budgetID int64,
) error {
	if importBatchID <= 0 {
		return nil
	}

	rawRowData, err := json.Marshal(buildRawRowData(header, rowValues))
	if err != nil {
		return apperror.Internal("failed to marshal raw import row data", err)
	}

	normalizedRowData, err := json.Marshal(buildNormalizedRowData(row))
	if err != nil {
		return apperror.Internal("failed to marshal normalized import row data", err)
	}

	messagesData, err := json.Marshal(messages)
	if err != nil {
		return apperror.Internal("failed to marshal import row messages", err)
	}

	_, err = s.auditRepo.CreateRowRaw(ctx, &model.BudgetImportRowRawModel{
		ImportBatchID:     importBatchID,
		RowNumber:         row.rowNumber,
		BudgetNumber:      row.budgetNumber,
		Status:            status,
		Action:            action,
		RawRowData:        rawRowData,
		NormalizedRowData: normalizedRowData,
		Messages:          messagesData,
		BudgetID:          validNullInt64(budgetID),
		CreatedAt:         time.Now().UTC(),
	})
	if err != nil {
		return apperror.Internal("failed to store import raw row", err)
	}

	return nil
}

func buildRawRowData(header []string, rowValues []string) map[string]string {
	maxColumns := len(rowValues)
	if len(header) > maxColumns {
		maxColumns = len(header)
	}

	data := make(map[string]string, maxColumns)
	for column := 0; column < maxColumns; column++ {
		key := normalizeCellText(getCell(header, column))
		if key == "" {
			key = fmt.Sprintf("column_%d", column+1)
		}
		data[key] = getCell(rowValues, column)
	}

	return data
}

func buildNormalizedRowData(row normalizedBudgetImportRow) map[string]interface{} {
	data := map[string]interface{}{
		"row_number":        row.rowNumber,
		"budget_number":     row.budgetNumber,
		"year_budget":       row.yearBudget,
		"revision":          row.revision,
		"gross_value":       row.grossValue,
		"commission_value":  row.commissionValue,
		"area_m2":           row.areaM2,
		"status_name":       row.statusName,
		"priority_name":     row.priorityName,
		"installer_name":    row.installerName,
		"project_name":      row.projectName,
		"project_type_name": row.projectTypeName,
		"salesperson_name":  row.salespersonName,
		"contact_name":      row.contactName,
		"loss_reason_name":  row.lossReasonName,
		"competitor_name":   row.competitorName,
		"designer_name":     row.designerName,
		"specification":     row.specification,
		"current_follow_up": row.currentFollowUp,
		"warnings":          row.warnings,
		"sent_at":           row.sentAt.Format(time.RFC3339),
		"competitor_price":  nil,
	}
	if row.competitorPrice != nil {
		data["competitor_price"] = *row.competitorPrice
	}

	return data
}

func fallbackName(raw string) string {
	if isMissingValue(raw) {
		return notInformedName
	}
	return normalizeCellText(raw)
}

func shouldSkipProjectAssociation(projectName string) bool {
	return normalizeLookupKey(projectName) == normalizeLookupKey(notInformedName)
}

func (s *service) loadCatalogRuntime(ctx context.Context) (*catalogRuntime, error) {
	statuses, err := s.statusRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to load budget statuses", err)
	}
	priorities, err := s.priorityRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to load priorities", err)
	}
	installers, err := s.installerRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to load installers", err)
	}
	projects, err := s.projectRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to load projects", err)
	}
	projectTypes, err := s.projectTypeRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to load project types", err)
	}
	salespeople, err := s.salespersonRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to load salespeople", err)
	}
	contacts, err := s.contactRepo.List(ctx, nil)
	if err != nil {
		return nil, apperror.Internal("failed to load contacts", err)
	}
	lossReasons, err := s.lossReasonRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to load loss reasons", err)
	}

	runtime := &catalogRuntime{
		statuses:     make(map[string]int64, len(statuses)),
		priorities:   make(map[string]int64, len(priorities)),
		installers:   make(map[string]int64, len(installers)),
		projects:     make(map[string]int64, len(projects)),
		projectTypes: make(map[string]int64, len(projectTypes)),
		salespeople:  make(map[string]int64, len(salespeople)),
		contacts:     make(map[string]int64, len(contacts)),
		lossReasons:  make(map[string]int64, len(lossReasons)),
	}

	for _, item := range statuses {
		runtime.statuses[normalizeLookupKey(item.Name)] = item.ID
	}
	for _, item := range priorities {
		runtime.priorities[normalizeLookupKey(item.Name)] = item.ID
	}
	for _, item := range installers {
		runtime.installers[normalizeLookupKey(item.Name)] = item.ID
	}
	for _, item := range projects {
		runtime.projects[normalizeLookupKey(item.Name)] = item.ID
	}
	for _, item := range projectTypes {
		runtime.projectTypes[normalizeLookupKey(item.Name)] = item.ID
	}
	for _, item := range salespeople {
		runtime.salespeople[normalizeLookupKey(item.Name)] = item.ID
	}
	for _, item := range contacts {
		key := buildContactRuntimeKey(item.InstallerID, item.Name)
		runtime.contacts[key] = item.ID
	}
	for _, item := range lossReasons {
		runtime.lossReasons[normalizeLookupKey(item.Name)] = item.ID
	}

	return runtime, nil
}

func (s *service) ensureBudgetStatusID(ctx context.Context, catalogs *catalogRuntime, name string, options dto.PreviewBudgetImportOptions) (int64, int, error) {
	key := normalizeLookupKey(name)
	if id, ok := catalogs.statuses[key]; ok {
		return id, 0, nil
	}
	if !options.CreateMissingCatalogs {
		return 0, 0, apperror.BadRequest("Status de orcamento nao encontrado para importacao")
	}

	now := time.Now()
	id, err := s.statusRepo.Create(ctx, &model.BudgetStatusModel{
		Code:        buildCatalogCode(name),
		Name:        name,
		Description: "Criado automaticamente pela importacao",
		IsFinal:     false,
		SortOrder:   999,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		existing, getErr := s.statusRepo.GetByCodeOrName(ctx, buildCatalogCode(name), name)
		if getErr == nil && existing != nil {
			catalogs.statuses[key] = existing.ID
			return existing.ID, 0, nil
		}
		return 0, 0, apperror.Internal("Falha ao criar status de orcamento para importacao", err)
	}

	catalogs.statuses[key] = id
	return id, 1, nil
}

func (s *service) ensurePriorityID(ctx context.Context, catalogs *catalogRuntime, name string, options dto.PreviewBudgetImportOptions) (int64, int, error) {
	key := normalizeLookupKey(name)
	if id, ok := catalogs.priorities[key]; ok {
		return id, 0, nil
	}
	if !options.CreateMissingCatalogs {
		return 0, 0, apperror.BadRequest("Prioridade nao encontrada para importacao")
	}

	now := time.Now()
	id, err := s.priorityRepo.Create(ctx, &model.PriorityModel{
		Code:      buildCatalogCode(name),
		Name:      name,
		Weight:    0,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		existing, getErr := s.priorityRepo.GetByCodeOrName(ctx, buildCatalogCode(name), name)
		if getErr == nil && existing != nil {
			catalogs.priorities[key] = existing.ID
			return existing.ID, 0, nil
		}
		return 0, 0, apperror.Internal("Falha ao criar prioridade para importacao", err)
	}

	catalogs.priorities[key] = id
	return id, 1, nil
}

func (s *service) ensureProjectTypeID(ctx context.Context, catalogs *catalogRuntime, name string, options dto.PreviewBudgetImportOptions) (int64, int, error) {
	key := normalizeLookupKey(name)
	if id, ok := catalogs.projectTypes[key]; ok {
		return id, 0, nil
	}
	if !options.CreateMissingCatalogs {
		return 0, 0, apperror.BadRequest("Tipo de projeto nao encontrado para importacao")
	}

	now := time.Now()
	id, err := s.projectTypeRepo.Create(ctx, &model.ProjectTypeModel{
		Code:        buildCatalogCode(name),
		Name:        name,
		Description: "Criado automaticamente pela importacao",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		existing, getErr := s.projectTypeRepo.GetByCodeOrName(ctx, buildCatalogCode(name), name)
		if getErr == nil && existing != nil {
			catalogs.projectTypes[key] = existing.ID
			return existing.ID, 0, nil
		}
		return 0, 0, apperror.Internal("Falha ao criar tipo de projeto para importacao", err)
	}

	catalogs.projectTypes[key] = id
	return id, 1, nil
}

func (s *service) ensureInstallerID(ctx context.Context, catalogs *catalogRuntime, name string, options dto.PreviewBudgetImportOptions) (int64, int, error) {
	key := normalizeLookupKey(name)
	if id, ok := catalogs.installers[key]; ok {
		return id, 0, nil
	}
	if !options.CreateMissingCatalogs {
		return 0, 0, apperror.BadRequest("Instalador nao encontrado para importacao")
	}

	now := time.Now()
	id, err := s.installerRepo.Create(ctx, &model.InstallerModel{
		Name:      name,
		Email:     buildCatalogEmail("installer", name),
		Phone:     buildCatalogPhone("1", name),
		City:      notInformedName,
		State:     "NI",
		Notes:     "Criado automaticamente pela importacao",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return 0, 0, apperror.Internal("Falha ao criar instalador para importacao", err)
	}

	catalogs.installers[key] = id
	return id, 1, nil
}

func (s *service) ensureProjectID(ctx context.Context, catalogs *catalogRuntime, name string, projectTypeID int64, options dto.PreviewBudgetImportOptions) (int64, int, error) {
	key := normalizeLookupKey(name)
	if id, ok := catalogs.projects[key]; ok {
		return id, 0, nil
	}
	if !options.CreateMissingCatalogs {
		return 0, 0, apperror.BadRequest("Projeto nao encontrado para importacao")
	}

	now := time.Now()
	id, err := s.projectRepo.Create(ctx, &model.ProjectModel{
		Name: name,
		ProjectTypeID: sql.NullInt64{
			Int64: projectTypeID,
			Valid: projectTypeID > 0,
		},
		City:      notInformedName,
		State:     "NI",
		Notes:     "Criado automaticamente pela importacao",
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return 0, 0, apperror.Internal("Falha ao criar projeto para importacao", err)
	}

	catalogs.projects[key] = id
	return id, 1, nil
}

func (s *service) ensureSalespersonID(ctx context.Context, catalogs *catalogRuntime, name string, options dto.PreviewBudgetImportOptions) (int64, int, error) {
	key := normalizeLookupKey(name)
	if id, ok := catalogs.salespeople[key]; ok {
		return id, 0, nil
	}
	if !options.CreateMissingCatalogs {
		return 0, 0, apperror.BadRequest("Vendedor nao encontrado para importacao")
	}

	now := time.Now()
	id, err := s.salespersonRepo.Create(ctx, &model.SalespersonModel{
		Name:      name,
		Email:     buildCatalogEmail("salesperson", name),
		Phone:     buildCatalogPhone("2", name),
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return 0, 0, apperror.Internal("Falha ao criar vendedor para importacao", err)
	}

	catalogs.salespeople[key] = id
	return id, 1, nil
}

func (s *service) ensureContactID(ctx context.Context, catalogs *catalogRuntime, installerID int64, name string, options dto.PreviewBudgetImportOptions) (int64, int, error) {
	key := buildContactRuntimeKey(installerID, name)
	if id, ok := catalogs.contacts[key]; ok {
		return id, 0, nil
	}
	if !options.CreateMissingCatalogs {
		return 0, 0, apperror.BadRequest("Contato nao encontrado para importacao")
	}

	now := time.Now()
	id, err := s.contactRepo.Create(ctx, &model.ContactModel{
		InstallerID: installerID,
		Name:        name,
		Email:       buildContactEmail(installerID, name),
		Phone:       buildContactPhone(installerID, name),
		Role:        notInformedName,
		IsPrimary:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return 0, 0, apperror.Internal("Falha ao criar contato para importacao", err)
	}

	catalogs.contacts[key] = id
	return id, 1, nil
}

func (s *service) ensureLossReasonID(ctx context.Context, catalogs *catalogRuntime, name string, options dto.PreviewBudgetImportOptions) (int64, int, error) {
	key := normalizeLookupKey(name)
	if id, ok := catalogs.lossReasons[key]; ok {
		return id, 0, nil
	}
	if !options.CreateMissingCatalogs {
		return 0, 0, apperror.BadRequest("Motivo de perda nao encontrado para importacao")
	}

	now := time.Now()
	id, err := s.lossReasonRepo.Create(ctx, &model.LossReasonModel{
		Code:        buildCatalogCode(name),
		Name:        name,
		Description: "Criado automaticamente pela importacao",
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		existing, getErr := s.lossReasonRepo.GetByCodeOrName(ctx, buildCatalogCode(name), name)
		if getErr == nil && existing != nil {
			catalogs.lossReasons[key] = existing.ID
			return existing.ID, 0, nil
		}
		return 0, 0, apperror.Internal("Falha ao criar motivo de perda para importacao", err)
	}

	catalogs.lossReasons[key] = id
	return id, 1, nil
}

func buildCatalogCode(name string) string {
	normalized := normalizeLookupKey(name)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	builder := strings.Builder{}
	for _, char := range normalized {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			builder.WriteRune(char)
		}
	}
	if builder.Len() == 0 {
		return "nao_informado"
	}
	return builder.String()
}

func buildCatalogEmail(prefix string, name string) string {
	return fmt.Sprintf("%s-%d@import.local", prefix, deterministicNumber(name))
}

func buildContactEmail(installerID int64, name string) string {
	return fmt.Sprintf("contact-%d-%d@import.local", installerID, deterministicNumber(name))
}

func buildCatalogPhone(prefix string, name string) string {
	return fmt.Sprintf("%s%010d", prefix, deterministicNumber(name)%10000000000)
}

func buildContactPhone(installerID int64, name string) string {
	return fmt.Sprintf("9%010d", (deterministicNumber(name)+int(installerID))%10000000000)
}

func deterministicNumber(value string) int {
	normalized := normalizeLookupKey(value)
	total := 0
	for _, char := range normalized {
		total += int(char)
	}
	if total == 0 {
		return 1
	}
	return total
}

func buildContactRuntimeKey(installerID int64, name string) string {
	return strconv.FormatInt(installerID, 10) + "|" + normalizeLookupKey(name)
}

func validNullInt64(value int64) sql.NullInt64 {
	if value <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: value, Valid: true}
}

func validNullFloat64(value *float64) sql.NullFloat64 {
	if value == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *value, Valid: true}
}
