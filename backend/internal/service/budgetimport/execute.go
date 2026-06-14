package budgetimport

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type previewSnapshot struct {
	previewID string
	fileName  string
	fileData  []byte
	options   dto.PreviewBudgetImportOptions
	createdAt time.Time
}

type importBudgetRow struct {
	rowNumber       int
	budgetNumber    string
	yearBudget      int
	revision        int
	sentAt          time.Time
	grossValue      float64
	commissionValue float64
	areaM2          float64
	statusName      string
	priorityName    string
	installerName   string
	projectName     string
	projectTypeName string
	salespersonName string
	contactName     string
	lossReasonName  string
	competitorName  string
	competitorPrice *float64
	designerName    string
	specification   string
	currentFollowUp string
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

	workbook, err := parseWorkbook(snapshot.fileData, previewSheetName)
	if err != nil {
		return nil, err
	}
	header := workbook.rows[previewHeaderRowNumber]
	if !isExpectedHeader(header) {
		return nil, apperror.BadRequest("Cabecalho invalido na aba ORCAMENTOS")
	}

	catalogs, err := s.loadCatalogRuntime(ctx)
	if err != nil {
		return nil, err
	}

	startedAt := time.Now().UTC()
	summary := dto.ExecuteBudgetImportSummary{}
	status := "completed"

	for rowNumber := previewHeaderRowNumber + 1; rowNumber <= workbook.maxRow; rowNumber++ {
		rowValues := workbook.rows[rowNumber]
		if isRowEmpty(rowValues) {
			continue
		}

		summary.RowsProcessed++

		row, rowErr := parseImportBudgetRow(rowNumber, rowValues)
		if rowErr != nil {
			summary.RowsFailed++
			status = "completed_with_errors"
			continue
		}

		rowResult, importErr := s.importBudgetRow(ctx, catalogs, snapshot.options, row)
		if importErr != nil {
			summary.RowsFailed++
			status = "completed_with_errors"
			continue
		}

		summary.CatalogsCreated += rowResult.catalogsCreated
		switch rowResult.action {
		case "create":
			summary.BudgetsCreated++
		case "update":
			summary.BudgetsUpdated++
		case "ignore":
			summary.BudgetsIgnored++
		}
	}

	finishedAt := time.Now().UTC()
	return &dto.ExecuteBudgetImportResponse{
		ImportID:   fmt.Sprintf("import_%d", finishedAt.UnixNano()),
		PreviewID:  snapshot.previewID,
		Status:     status,
		StartedAt:  startedAt.Format(time.RFC3339),
		FinishedAt: finishedAt.Format(time.RFC3339),
		Summary:    summary,
		Result: dto.ExecuteBudgetImportResult{
			Message: "Importacao concluida com sucesso.",
		},
	}, nil
}

type importRowResult struct {
	action          string
	catalogsCreated int
}

func (s *service) importBudgetRow(
	ctx context.Context,
	catalogs *catalogRuntime,
	options dto.PreviewBudgetImportOptions,
	row importBudgetRow,
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

	projectTypeID, created, err := s.ensureProjectTypeID(ctx, catalogs, row.projectTypeName, options)
	if err != nil {
		return importRowResult{}, err
	}
	catalogsCreated += created

	installerID, created, err := s.ensureInstallerID(ctx, catalogs, row.installerName, options)
	if err != nil {
		return importRowResult{}, err
	}
	catalogsCreated += created

	projectID, created, err := s.ensureProjectID(ctx, catalogs, row.projectName, projectTypeID, options)
	if err != nil {
		return importRowResult{}, err
	}
	catalogsCreated += created

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

	existingBudget, err := s.budgetRepo.GetByNumberAndYear(ctx, row.budgetNumber, row.yearBudget)
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
		UpdatedAt:            time.Now(),
	}

	if existingBudget != nil {
		if options.DuplicateStrategy == duplicateStrategyIgnore {
			return importRowResult{
				action:          "ignore",
				catalogsCreated: catalogsCreated,
			}, nil
		}

		budgetModel.ID = existingBudget.ID
		if err := s.budgetRepo.Update(ctx, budgetModel); err != nil {
			return importRowResult{}, apperror.Internal("failed to update budget", err)
		}

		return importRowResult{
			action:          "update",
			catalogsCreated: catalogsCreated,
		}, nil
	}

	budgetModel.CreatedAt = budgetModel.UpdatedAt
	if _, err := s.budgetRepo.Create(ctx, budgetModel); err != nil {
		return importRowResult{}, apperror.Internal("failed to create budget", err)
	}

	return importRowResult{
		action:          "create",
		catalogsCreated: catalogsCreated,
	}, nil
}

func parseImportBudgetRow(rowNumber int, rowValues []string) (importBudgetRow, error) {
	budgetNumber := normalizeCellText(getCell(rowValues, 1))
	if budgetNumber == "" {
		return importBudgetRow{}, fmt.Errorf("budget number missing")
	}

	sentAt, err := parseExcelDate(getCell(rowValues, 0))
	if err != nil {
		return importBudgetRow{}, err
	}

	grossValue, err := parseOptionalNumber(getCell(rowValues, 8), true)
	if err != nil {
		return importBudgetRow{}, err
	}
	commissionValue, err := parseOptionalNumber(getCell(rowValues, 9), false)
	if err != nil {
		return importBudgetRow{}, err
	}
	areaM2, err := parseOptionalNumber(getCell(rowValues, 10), false)
	if err != nil {
		return importBudgetRow{}, err
	}

	var competitorPrice *float64
	if !isMissingValue(getCell(rowValues, 15)) {
		value, valueErr := parseOptionalNumber(getCell(rowValues, 15), false)
		if valueErr != nil {
			return importBudgetRow{}, valueErr
		}
		competitorPrice = &value
	}

	return importBudgetRow{
		rowNumber:       rowNumber,
		budgetNumber:    budgetNumber,
		yearBudget:      sentAt.Year(),
		revision:        extractRevision(getCell(rowValues, 2)),
		sentAt:          sentAt,
		grossValue:      grossValue,
		commissionValue: commissionValue,
		areaM2:          areaM2,
		statusName:      fallbackName(getCell(rowValues, 11)),
		priorityName:    notInformedName,
		installerName:   fallbackName(getCell(rowValues, 3)),
		projectName:     fallbackName(getCell(rowValues, 4)),
		projectTypeName: fallbackName(getCell(rowValues, 5)),
		salespersonName: fallbackName(getCell(rowValues, 6)),
		contactName:     fallbackName(getCell(rowValues, 7)),
		lossReasonName:  fallbackName(getCell(rowValues, 14)),
		competitorName:  fallbackName(getCell(rowValues, 13)),
		competitorPrice: competitorPrice,
		designerName:    fallbackName(getCell(rowValues, 16)),
		specification:   fallbackName(getCell(rowValues, 17)),
		currentFollowUp: fallbackName(getCell(rowValues, 12)),
	}, nil
}

func fallbackName(raw string) string {
	if isMissingValue(raw) {
		return notInformedName
	}
	return normalizeCellText(raw)
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
