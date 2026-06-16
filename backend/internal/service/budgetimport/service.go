package budgetimport

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

const (
	notInformedName         = "Nao informado"
	duplicateStrategyIgnore = "ignore"
	duplicateStrategyUpdate = "update"
)

var revisionDigitsPattern = regexp.MustCompile(`\d+`)

type budgetRepository interface {
	Create(ctx context.Context, item *model.BudgetModel) (int64, error)
	ExistsByNumberAndYear(ctx context.Context, budgetNumber string, yearBudget int) (bool, error)
	ExistsBySourceAndNumberAndYear(ctx context.Context, sourceCompany string, budgetNumber string, yearBudget int) (bool, error)
	GetByNumberAndYear(ctx context.Context, budgetNumber string, yearBudget int) (*model.BudgetModel, error)
	GetBySourceAndNumberAndYear(ctx context.Context, sourceCompany string, budgetNumber string, yearBudget int) (*model.BudgetModel, error)
	Update(ctx context.Context, item *model.BudgetModel) error
}

type budgetStatusRepository interface {
	Create(ctx context.Context, status *model.BudgetStatusModel) (int64, error)
	GetByCodeOrName(ctx context.Context, code string, name string) (*model.BudgetStatusModel, error)
	List(ctx context.Context) ([]model.BudgetStatusModel, error)
}

type priorityRepository interface {
	Create(ctx context.Context, priority *model.PriorityModel) (int64, error)
	GetByCodeOrName(ctx context.Context, code string, name string) (*model.PriorityModel, error)
	List(ctx context.Context) ([]model.PriorityModel, error)
}

type installerRepository interface {
	Create(ctx context.Context, installer *model.InstallerModel) (int64, error)
	List(ctx context.Context) ([]model.InstallerModel, error)
}

type projectRepository interface {
	Create(ctx context.Context, item *model.ProjectModel) (int64, error)
	List(ctx context.Context) ([]model.ProjectModel, error)
}

type projectTypeRepository interface {
	Create(ctx context.Context, item *model.ProjectTypeModel) (int64, error)
	GetByCodeOrName(ctx context.Context, code string, name string) (*model.ProjectTypeModel, error)
	List(ctx context.Context) ([]model.ProjectTypeModel, error)
}

type salespersonRepository interface {
	Create(ctx context.Context, salesperson *model.SalespersonModel) (int64, error)
	List(ctx context.Context) ([]model.SalespersonModel, error)
}

type contactRepository interface {
	Create(ctx context.Context, contact *model.ContactModel) (int64, error)
	List(ctx context.Context, installerID *int64) ([]model.ContactModel, error)
}

type lossReasonRepository interface {
	Create(ctx context.Context, reason *model.LossReasonModel) (int64, error)
	GetByCodeOrName(ctx context.Context, code string, name string) (*model.LossReasonModel, error)
	List(ctx context.Context) ([]model.LossReasonModel, error)
}

type importAuditRepository interface {
	CreateBatch(ctx context.Context, item *model.BudgetImportBatchModel) (int64, error)
	UpdateBatch(ctx context.Context, item *model.BudgetImportBatchModel) error
	GetBatchByID(ctx context.Context, batchID int64) (*model.BudgetImportBatchModel, error)
	CreateRowRaw(ctx context.Context, item *model.BudgetImportRowRawModel) (int64, error)
}

type Service interface {
	Preview(
		ctx context.Context,
		fileName string,
		fileData []byte,
		options dto.PreviewBudgetImportOptions,
	) (*dto.PreviewBudgetImportResponse, error)
	StartImport(ctx context.Context, req *dto.ExecuteBudgetImportRequest) (*dto.ExecuteBudgetImportResponse, error)
	ExecuteImport(ctx context.Context, req *dto.ExecuteBudgetImportRequest) (*dto.ExecuteBudgetImportResponse, error)
	GetImportStatus(ctx context.Context, importBatchID int64) (*dto.ExecuteBudgetImportResponse, error)
}

type service struct {
	budgetRepo      budgetRepository
	statusRepo      budgetStatusRepository
	priorityRepo    priorityRepository
	installerRepo   installerRepository
	projectRepo     projectRepository
	projectTypeRepo projectTypeRepository
	salespersonRepo salespersonRepository
	contactRepo     contactRepository
	lossReasonRepo  lossReasonRepository
	auditRepo       importAuditRepository
	mu              sync.Mutex
	previews        map[string]previewSnapshot
}

type workbookData struct {
	rows   map[int][]string
	maxRow int
}

type sheetRowXML struct {
	Number int            `xml:"r,attr"`
	Cells  []sheetCellXML `xml:"c"`
}

type sheetCellXML struct {
	Reference string `xml:"r,attr"`
	Type      string `xml:"t,attr"`
	Value     string `xml:"v"`
	Inline    string `xml:"is>t"`
}

type sharedStringsXML struct {
	Items []sharedStringItemXML `xml:"si"`
}

type sharedStringItemXML struct {
	Text string               `xml:"t"`
	Runs []sharedStringRunXML `xml:"r"`
}

type sharedStringRunXML struct {
	Text string `xml:"t"`
}

type workbookXML struct {
	Sheets []workbookSheetXML `xml:"sheets>sheet"`
}

type workbookSheetXML struct {
	Name           string `xml:"name,attr"`
	RelationshipID string `xml:"id,attr"`
}

type workbookRelationshipsXML struct {
	Items []workbookRelationshipXML `xml:"Relationship"`
}

type workbookRelationshipXML struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type catalogIndex struct {
	statuses                 map[string]struct{}
	priorities               map[string]struct{}
	installers               map[string]struct{}
	projects                 map[string]struct{}
	projectTypes             map[string]struct{}
	salespeople              map[string]struct{}
	contacts                 map[string]struct{}
	lossReasons              map[string]struct{}
	installerNameByID        map[int64]string
	defaultStatusExists      bool
	defaultPriorityExists    bool
	defaultInstallerExists   bool
	defaultProjectExists     bool
	defaultProjectTypeExists bool
	defaultSalespersonExists bool
	defaultContactExists     bool
	defaultLossReasonExists  bool
}

func NewService(
	budgetRepo budgetRepository,
	statusRepo budgetStatusRepository,
	priorityRepo priorityRepository,
	installerRepo installerRepository,
	projectRepo projectRepository,
	projectTypeRepo projectTypeRepository,
	salespersonRepo salespersonRepository,
	contactRepo contactRepository,
	lossReasonRepo lossReasonRepository,
	auditRepo importAuditRepository,
) Service {
	if auditRepo == nil {
		auditRepo = noopImportAuditRepository{}
	}

	return &service{
		budgetRepo:      budgetRepo,
		statusRepo:      statusRepo,
		priorityRepo:    priorityRepo,
		installerRepo:   installerRepo,
		projectRepo:     projectRepo,
		projectTypeRepo: projectTypeRepo,
		salespersonRepo: salespersonRepo,
		contactRepo:     contactRepo,
		lossReasonRepo:  lossReasonRepo,
		auditRepo:       auditRepo,
		previews:        make(map[string]previewSnapshot),
	}
}

type noopImportAuditRepository struct{}

func (noopImportAuditRepository) CreateBatch(_ context.Context, _ *model.BudgetImportBatchModel) (int64, error) {
	return 0, nil
}

func (noopImportAuditRepository) UpdateBatch(_ context.Context, _ *model.BudgetImportBatchModel) error {
	return nil
}

func (noopImportAuditRepository) GetBatchByID(_ context.Context, _ int64) (*model.BudgetImportBatchModel, error) {
	return nil, nil
}

func (noopImportAuditRepository) CreateRowRaw(_ context.Context, _ *model.BudgetImportRowRawModel) (int64, error) {
	return 0, nil
}

func validNullTime(value time.Time) sql.NullTime {
	if value.IsZero() {
		return sql.NullTime{}
	}

	return sql.NullTime{
		Time:  value,
		Valid: true,
	}
}

func (s *service) Preview(
	ctx context.Context,
	fileName string,
	fileData []byte,
	options dto.PreviewBudgetImportOptions,
) (*dto.PreviewBudgetImportResponse, error) {
	if len(fileData) == 0 {
		return nil, apperror.BadRequest("Arquivo obrigatorio")
	}

	if strings.ToLower(filepath.Ext(fileName)) != ".xlsx" {
		return nil, apperror.BadRequest("Apenas arquivos .xlsx sao suportados")
	}

	normalizedOptions, err := normalizePreviewOptions(options)
	if err != nil {
		return nil, err
	}

	layout, workbook, err := resolveImportLayout(fileData)
	if err != nil {
		return nil, err
	}
	importLogger := logger.FromContext(ctx).With(
		slog.String("import_action", "preview"),
		slog.String("source_layout", layout.Key()),
		slog.String("source_company", layout.SourceCompany()),
		slog.String("file_name", fileName),
	)
	importLogger.InfoContext(ctx, "preview de importacao iniciado")

	catalogs, err := s.loadCatalogIndex(ctx)
	if err != nil {
		return nil, err
	}

	response := &dto.PreviewBudgetImportResponse{
		PreviewID: fmt.Sprintf("preview_%d", time.Now().UnixNano()),
		FileName:  fileName,
		SheetName: layout.SheetName(),
		HeaderRow: layout.HeaderRowNumber(),
		Layout: dto.BudgetImportPreviewLayoutInfo{
			Key:           layout.Key(),
			Name:          layout.Name(),
			SourceCompany: layout.SourceCompany(),
			Description:   layout.Description(),
		},
		FieldGroups:       clonePreviewFieldGroups(layout.FieldGroups()),
		Governance:        clonePreviewGovernance(layout.Governance()),
		Options:           normalizedOptions,
		Warnings:          clonePreviewWarnings(layout.PreviewWarnings()),
		Errors:            []dto.BudgetImportPreviewMessage{},
		SampleRows:        make([]dto.BudgetImportPreviewRow, 0, 20),
		InconsistencyRows: make([]dto.BudgetImportPreviewRow, 0),
	}

	missingStatuses := make(map[string]struct{})
	missingPriorities := make(map[string]struct{})
	missingInstallers := make(map[string]struct{})
	missingProjects := make(map[string]struct{})
	missingProjectTypes := make(map[string]struct{})
	missingSalespeople := make(map[string]struct{})
	missingContacts := make(map[string]struct{})
	missingLossReasons := make(map[string]struct{})
	fileDuplicateKeys := make(map[string]struct{})

	ensureDefaultCatalog(
		catalogs.defaultStatusExists,
		missingStatuses,
		normalizedOptions,
	)
	ensureDefaultCatalog(
		catalogs.defaultPriorityExists,
		missingPriorities,
		normalizedOptions,
	)
	ensureDefaultCatalog(
		catalogs.defaultInstallerExists,
		missingInstallers,
		normalizedOptions,
	)
	ensureDefaultCatalog(
		catalogs.defaultProjectExists,
		missingProjects,
		normalizedOptions,
	)
	ensureDefaultCatalog(
		catalogs.defaultProjectTypeExists,
		missingProjectTypes,
		normalizedOptions,
	)
	ensureDefaultCatalog(
		catalogs.defaultSalespersonExists,
		missingSalespeople,
		normalizedOptions,
	)
	ensureDefaultCatalog(
		catalogs.defaultContactExists,
		missingContacts,
		normalizedOptions,
	)
	ensureDefaultCatalog(
		catalogs.defaultLossReasonExists,
		missingLossReasons,
		normalizedOptions,
	)

	for rowNumber := layout.HeaderRowNumber() + 1; rowNumber <= workbook.maxRow; rowNumber++ {
		rowValues := workbook.rows[rowNumber]
		if layout.IsRowEmpty(rowValues) {
			response.Summary.RowsEmptyIgnored++
			continue
		}

		response.Summary.RowsRead++

		rowPreview, result, rowErr := s.buildRowPreview(
			ctx,
			layout,
			rowNumber,
			rowValues,
			catalogs,
			normalizedOptions,
			fileDuplicateKeys,
			missingStatuses,
			missingPriorities,
			missingInstallers,
			missingProjects,
			missingProjectTypes,
			missingSalespeople,
			missingContacts,
			missingLossReasons,
		)
		if rowErr != nil {
			return nil, rowErr
		}

		if len(response.SampleRows) < 20 {
			response.SampleRows = append(response.SampleRows, rowPreview)
		}

		if result.hasError {
			response.InconsistencyRows = append(response.InconsistencyRows, rowPreview)
			response.Summary.RowsWithError++
			continue
		}

		response.Summary.RowsValid++
		if result.hasWarning {
			response.Summary.RowsWithWarning++
		}
		if result.isExistingBudget {
			response.Summary.ExistingBudgets++
		} else {
			response.Summary.NewBudgets++
		}
	}

	if response.Summary.RowsRead == 0 {
		return nil, apperror.BadRequest(fmt.Sprintf("A aba %s nao contem linhas importaveis", layout.SheetName()))
	}

	response.CatalogActions = dto.BudgetImportCatalogActions{
		BudgetStatusesToCreate: len(missingStatuses),
		PrioritiesToCreate:     len(missingPriorities),
		InstallersToCreate:     len(missingInstallers),
		ProjectsToCreate:       len(missingProjects),
		ProjectTypesToCreate:   len(missingProjectTypes),
		SalespeopleToCreate:    len(missingSalespeople),
		ContactsToCreate:       len(missingContacts),
		LossReasonsToCreate:    len(missingLossReasons),
	}

	s.storePreview(previewSnapshot{
		previewID: response.PreviewID,
		fileName:  fileName,
		fileData:  append([]byte(nil), fileData...),
		layoutKey: layout.Key(),
		options:   normalizedOptions,
	})

	importLogger.InfoContext(
		ctx,
		"preview de importacao concluido",
		slog.String("preview_id", response.PreviewID),
		slog.Int("rows_read", response.Summary.RowsRead),
		slog.Int("rows_valid", response.Summary.RowsValid),
		slog.Int("rows_with_warning", response.Summary.RowsWithWarning),
		slog.Int("rows_with_error", response.Summary.RowsWithError),
		slog.Int("new_budgets", response.Summary.NewBudgets),
		slog.Int("existing_budgets", response.Summary.ExistingBudgets),
	)

	return response, nil
}

type rowPreviewResult struct {
	hasWarning       bool
	hasError         bool
	isExistingBudget bool
}

func (s *service) buildRowPreview(
	ctx context.Context,
	layout importLayout,
	rowNumber int,
	rowValues []string,
	catalogs *catalogIndex,
	options dto.PreviewBudgetImportOptions,
	fileDuplicateKeys map[string]struct{},
	missingStatuses map[string]struct{},
	missingPriorities map[string]struct{},
	missingInstallers map[string]struct{},
	missingProjects map[string]struct{},
	missingProjectTypes map[string]struct{},
	missingSalespeople map[string]struct{},
	missingContacts map[string]struct{},
	missingLossReasons map[string]struct{},
) (dto.BudgetImportPreviewRow, rowPreviewResult, error) {
	normalizedRow, parseErr := layout.ParseNormalizedRow(rowNumber, rowValues)
	if parseErr != nil {
		return markRowError(dto.BudgetImportPreviewRow{
			RowNumber:    rowNumber,
			BudgetNumber: normalizedRow.budgetNumber,
			Messages:     []string{},
		}, parseErr.Error()), resultWithError(), nil
	}

	previewRow := dto.BudgetImportPreviewRow{
		RowNumber:    rowNumber,
		BudgetNumber: normalizedRow.budgetNumber,
		Messages:     append([]string{}, normalizedRow.warnings...),
	}
	result := rowPreviewResult{}
	result.hasWarning = len(normalizedRow.warnings) > 0

	installerName, installerHasWarning, installerHasError := resolveCatalogName(
		normalizedRow.installerName,
		"Instalador",
		catalogs.installers,
		catalogs.defaultInstallerExists,
		options,
		missingInstallers,
		&previewRow.Messages,
	)
	result.hasWarning = result.hasWarning || installerHasWarning
	result.hasError = result.hasError || installerHasError

	projectTypeName, projectTypeHasWarning, projectTypeHasError := resolveCatalogName(
		normalizedRow.projectTypeName,
		"Tipo de obra",
		catalogs.projectTypes,
		catalogs.defaultProjectTypeExists,
		options,
		missingProjectTypes,
		&previewRow.Messages,
	)
	result.hasWarning = result.hasWarning || projectTypeHasWarning
	result.hasError = result.hasError || projectTypeHasError

	_ = projectTypeName

	projectName, projectHasWarning, projectHasError := resolveCatalogName(
		normalizedRow.projectName,
		"Projeto",
		catalogs.projects,
		catalogs.defaultProjectExists,
		options,
		missingProjects,
		&previewRow.Messages,
	)
	result.hasWarning = result.hasWarning || projectHasWarning
	result.hasError = result.hasError || projectHasError

	_ = projectName

	_, salespersonHasWarning, salespersonHasError := resolveSalespersonName(
		normalizedRow.salespersonName,
		catalogs.salespeople,
		catalogs.defaultSalespersonExists,
		options,
		missingSalespeople,
		&previewRow.Messages,
	)
	result.hasWarning = result.hasWarning || salespersonHasWarning
	result.hasError = result.hasError || salespersonHasError

	_, statusHasWarning, statusHasError := resolveCatalogName(
		normalizedRow.statusName,
		"Status",
		catalogs.statuses,
		catalogs.defaultStatusExists,
		options,
		missingStatuses,
		&previewRow.Messages,
	)
	result.hasWarning = result.hasWarning || statusHasWarning
	result.hasError = result.hasError || statusHasError

	_, lossReasonHasWarning, lossReasonHasError := resolveCatalogName(
		normalizedRow.lossReasonName,
		"Motivo de perda",
		catalogs.lossReasons,
		catalogs.defaultLossReasonExists,
		options,
		missingLossReasons,
		&previewRow.Messages,
	)
	result.hasWarning = result.hasWarning || lossReasonHasWarning
	result.hasError = result.hasError || lossReasonHasError

	priorityHasWarning, priorityHasError := ensureDefaultPriority(
		catalogs.defaultPriorityExists,
		options,
		missingPriorities,
		&previewRow.Messages,
	)
	result.hasWarning = result.hasWarning || priorityHasWarning
	result.hasError = result.hasError || priorityHasError

	if installerName == "" {
		installerName = normalizeCellText(notInformedName)
	}
	_, contactHasWarning, contactHasError := resolveContactName(
		normalizedRow.contactName,
		installerName,
		catalogs.contacts,
		catalogs.defaultContactExists,
		options,
		missingContacts,
		&previewRow.Messages,
	)
	result.hasWarning = result.hasWarning || contactHasWarning
	result.hasError = result.hasError || contactHasError

	if result.hasError {
		previewRow.Status = "error"
		previewRow.Action = "error"
		return previewRow, result, nil
	}

	duplicateKey := fmt.Sprintf("%s|%d", normalizeLookupKey(normalizedRow.budgetNumber), normalizedRow.yearBudget)
	if _, exists := fileDuplicateKeys[duplicateKey]; exists {
		return markRowError(previewRow, "Orcamento duplicado dentro do arquivo para o mesmo ano."), resultWithError(), nil
	}
	fileDuplicateKeys[duplicateKey] = struct{}{}

	existsInDatabase, err := s.budgetRepo.ExistsBySourceAndNumberAndYear(
		ctx,
		layout.SourceCompany(),
		normalizedRow.budgetNumber,
		normalizedRow.yearBudget,
	)
	if err != nil {
		return dto.BudgetImportPreviewRow{}, rowPreviewResult{}, apperror.Internal("failed to check budget import preview", err)
	}

	result.isExistingBudget = existsInDatabase
	if existsInDatabase {
		if options.DuplicateStrategy == duplicateStrategyIgnore {
			previewRow.Action = "ignore"
		} else {
			previewRow.Action = "update"
		}
	} else {
		previewRow.Action = "create"
	}

	if result.hasWarning {
		previewRow.Status = "warning"
	} else {
		previewRow.Status = "ready"
	}

	return previewRow, result, nil
}

func normalizePreviewOptions(options dto.PreviewBudgetImportOptions) (dto.PreviewBudgetImportOptions, error) {
	normalized := options
	normalized.DuplicateStrategy = strings.ToLower(strings.TrimSpace(normalized.DuplicateStrategy))
	if normalized.DuplicateStrategy == "" {
		normalized.DuplicateStrategy = duplicateStrategyUpdate
	}
	if normalized.DuplicateStrategy != duplicateStrategyIgnore && normalized.DuplicateStrategy != duplicateStrategyUpdate {
		return dto.PreviewBudgetImportOptions{}, apperror.BadRequest("duplicate_strategy deve ser ignore ou update")
	}
	if !options.CreateMissingCatalogs && !options.UseDefaultNotInformed {
		normalized.CreateMissingCatalogs = false
		normalized.UseDefaultNotInformed = false
		return normalized, nil
	}
	if !options.CreateMissingCatalogs {
		normalized.CreateMissingCatalogs = false
	} else {
		normalized.CreateMissingCatalogs = true
	}
	if !options.UseDefaultNotInformed {
		normalized.UseDefaultNotInformed = false
	} else {
		normalized.UseDefaultNotInformed = true
	}
	return normalized, nil
}

func (s *service) loadCatalogIndex(ctx context.Context) (*catalogIndex, error) {
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

	index := &catalogIndex{
		statuses:          make(map[string]struct{}, len(statuses)),
		priorities:        make(map[string]struct{}, len(priorities)),
		installers:        make(map[string]struct{}, len(installers)),
		projects:          make(map[string]struct{}, len(projects)),
		projectTypes:      make(map[string]struct{}, len(projectTypes)),
		salespeople:       make(map[string]struct{}, len(salespeople)),
		contacts:          make(map[string]struct{}, len(contacts)),
		lossReasons:       make(map[string]struct{}, len(lossReasons)),
		installerNameByID: make(map[int64]string, len(installers)),
	}

	for _, item := range statuses {
		key := normalizeLookupKey(item.Name)
		index.statuses[key] = struct{}{}
		if key == normalizeLookupKey(notInformedName) {
			index.defaultStatusExists = true
		}
	}
	for _, item := range priorities {
		key := normalizeLookupKey(item.Name)
		index.priorities[key] = struct{}{}
		if key == normalizeLookupKey(notInformedName) {
			index.defaultPriorityExists = true
		}
	}
	for _, item := range installers {
		key := normalizeLookupKey(item.Name)
		index.installers[key] = struct{}{}
		index.installerNameByID[item.ID] = key
		if key == normalizeLookupKey(notInformedName) {
			index.defaultInstallerExists = true
		}
	}
	for _, item := range projects {
		key := normalizeLookupKey(item.Name)
		index.projects[key] = struct{}{}
		if key == normalizeLookupKey(notInformedName) {
			index.defaultProjectExists = true
		}
	}
	for _, item := range projectTypes {
		key := normalizeLookupKey(item.Name)
		index.projectTypes[key] = struct{}{}
		if key == normalizeLookupKey(notInformedName) {
			index.defaultProjectTypeExists = true
		}
	}
	for _, item := range salespeople {
		key := normalizeLookupKey(item.Name)
		if key == "" {
			continue
		}

		index.salespeople[key] = struct{}{}
		if key == normalizeLookupKey(notInformedName) {
			index.defaultSalespersonExists = true
		}
	}
	for aliasKey, canonicalID := range buildCanonicalSalespersonFirstNameIDs(salespeople) {
		if aliasKey == "" || canonicalID <= 0 {
			continue
		}

		index.salespeople[aliasKey] = struct{}{}
	}
	for aliasKey, aliasCount := range buildSalespersonFirstNameCounts(salespeople) {
		if aliasKey == "" || aliasCount != 1 {
			continue
		}

		index.salespeople[aliasKey] = struct{}{}
	}
	for _, item := range contacts {
		installerKey := index.installerNameByID[item.InstallerID]
		if installerKey == "" {
			installerKey = normalizeLookupKey(notInformedName)
		}
		key := installerKey + "|" + normalizeLookupKey(item.Name)
		index.contacts[key] = struct{}{}
		if normalizeLookupKey(item.Name) == normalizeLookupKey(notInformedName) {
			index.defaultContactExists = true
		}
	}
	for _, item := range lossReasons {
		key := normalizeLookupKey(item.Name)
		index.lossReasons[key] = struct{}{}
		if key == normalizeLookupKey(notInformedName) {
			index.defaultLossReasonExists = true
		}
	}

	return index, nil
}

func parseWorkbook(fileData []byte, sheetName string) (*workbookData, error) {
	reader, err := zip.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		return nil, apperror.BadRequest("Arquivo xlsx invalido")
	}

	sharedStrings, err := readSharedStrings(reader)
	if err != nil {
		return nil, apperror.BadRequest("Falha ao ler as strings compartilhadas do arquivo xlsx")
	}

	sheetPath, err := resolveSheetPath(reader, sheetName)
	if err != nil {
		return nil, err
	}

	sheetFile := findZipFile(reader, sheetPath)
	if sheetFile == nil {
		return nil, apperror.BadRequest("Dados da aba nao encontrados no arquivo xlsx")
	}

	rows, maxRow, err := readSheetRows(sheetFile, sharedStrings)
	if err != nil {
		return nil, apperror.BadRequest(fmt.Sprintf("Falha ao processar a aba %s", sheetName))
	}

	return &workbookData{
		rows:   rows,
		maxRow: maxRow,
	}, nil
}

func readSharedStrings(reader *zip.Reader) ([]string, error) {
	file := findZipFile(reader, "xl/sharedStrings.xml")
	if file == nil {
		return []string{}, nil
	}

	content, err := readZipFile(file)
	if err != nil {
		return nil, err
	}

	var parsed sharedStringsXML
	if err := xml.Unmarshal(content, &parsed); err != nil {
		return nil, err
	}

	items := make([]string, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		if item.Text != "" {
			items = append(items, item.Text)
			continue
		}

		builder := strings.Builder{}
		for _, run := range item.Runs {
			builder.WriteString(run.Text)
		}
		items = append(items, builder.String())
	}

	return items, nil
}

func resolveSheetPath(reader *zip.Reader, sheetName string) (string, error) {
	workbookFile := findZipFile(reader, "xl/workbook.xml")
	if workbookFile == nil {
		return "", apperror.BadRequest("Definicao da pasta de trabalho nao encontrada no arquivo xlsx")
	}
	workbookRelsFile := findZipFile(reader, "xl/_rels/workbook.xml.rels")
	if workbookRelsFile == nil {
		return "", apperror.BadRequest("Relacionamentos da pasta de trabalho nao encontrados no arquivo xlsx")
	}

	workbookContent, err := readZipFile(workbookFile)
	if err != nil {
		return "", err
	}
	workbookRelsContent, err := readZipFile(workbookRelsFile)
	if err != nil {
		return "", err
	}

	var workbook workbookXML
	if err := xml.Unmarshal(workbookContent, &workbook); err != nil {
		return "", err
	}
	var relationships workbookRelationshipsXML
	if err := xml.Unmarshal(workbookRelsContent, &relationships); err != nil {
		return "", err
	}

	relationshipByID := make(map[string]string, len(relationships.Items))
	for _, item := range relationships.Items {
		relationshipByID[item.ID] = item.Target
	}

	for _, sheet := range workbook.Sheets {
		if strings.EqualFold(strings.TrimSpace(sheet.Name), sheetName) {
			target := relationshipByID[sheet.RelationshipID]
			if target == "" {
				break
			}
			if strings.HasPrefix(target, "xl/") {
				return target, nil
			}
			return "xl/" + strings.TrimPrefix(target, "/"), nil
		}
	}

	return "", apperror.BadRequest(fmt.Sprintf("A aba %s nao foi encontrada no arquivo xlsx", sheetName))
}

func readSheetRows(file *zip.File, sharedStrings []string) (map[int][]string, int, error) {
	content, err := readZipFile(file)
	if err != nil {
		return nil, 0, err
	}

	decoder := xml.NewDecoder(bytes.NewReader(content))
	rows := make(map[int][]string)
	maxRow := 0
	for {
		token, tokenErr := decoder.Token()
		if tokenErr == io.EOF {
			break
		}
		if tokenErr != nil {
			return nil, 0, tokenErr
		}

		startElement, ok := token.(xml.StartElement)
		if !ok || startElement.Name.Local != "row" {
			continue
		}

		var row sheetRowXML
		if err := decoder.DecodeElement(&row, &startElement); err != nil {
			return nil, 0, err
		}

		if row.Number > maxRow {
			maxRow = row.Number
		}

		values := make([]string, 0, len(row.Cells))
		for _, cell := range row.Cells {
			index := len(values)
			if cell.Reference != "" {
				index = excelColumnIndex(cell.Reference)
			}
			if index < 0 {
				index = len(values)
			}
			for len(values) <= index {
				values = append(values, "")
			}
			values[index] = decodeCellValue(cell, sharedStrings)
		}
		rows[row.Number] = values
	}

	return rows, maxRow, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func findZipFile(reader *zip.Reader, name string) *zip.File {
	for _, file := range reader.File {
		if file.Name == name {
			return file
		}
	}
	return nil
}

func decodeCellValue(cell sheetCellXML, sharedStrings []string) string {
	if cell.Type == "inlineStr" {
		return cell.Inline
	}
	if cell.Type == "s" {
		index, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err == nil && index >= 0 && index < len(sharedStrings) {
			return sharedStrings[index]
		}
	}
	return strings.TrimSpace(cell.Value)
}

func excelColumnIndex(reference string) int {
	letters := strings.Builder{}
	for _, char := range reference {
		if unicode.IsLetter(char) {
			letters.WriteRune(unicode.ToUpper(char))
		}
	}
	value := letters.String()
	if value == "" {
		return -1
	}

	result := 0
	for _, char := range value {
		result = result*26 + int(char-'A'+1)
	}
	return result - 1
}

func getCell(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}
	return values[index]
}

func parseExcelDate(raw string) (time.Time, error) {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	serial, err := strconv.ParseFloat(strings.ReplaceAll(normalized, ",", "."), 64)
	if err != nil {
		return time.Time{}, err
	}

	baseDate := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	days := int(serial)
	fraction := serial - float64(days)
	duration := time.Duration(fraction * float64(24*time.Hour))

	return baseDate.AddDate(0, 0, days).Add(duration), nil
}

func parseOptionalNumber(raw string, required bool) (float64, error) {
	if isMissingValue(raw) {
		if required {
			return 0, fmt.Errorf("missing value")
		}
		return 0, nil
	}

	normalized := strings.ReplaceAll(strings.TrimSpace(raw), ",", ".")
	value, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func extractRevision(raw string) int {
	if isMissingValue(raw) {
		return 0
	}

	matches := revisionDigitsPattern.FindAllString(raw, -1)
	if len(matches) == 0 {
		return 0
	}

	lastMatch := matches[len(matches)-1]
	value, err := strconv.Atoi(lastMatch)
	if err != nil {
		return 0
	}
	return value
}

func isMissingValue(raw string) bool {
	normalized := strings.ToUpper(normalizeCellText(raw))
	switch normalized {
	case "", "-", "N/E", "N/I", "NULL":
		return true
	default:
		return false
	}
}

func normalizeCellText(raw string) string {
	parts := strings.Fields(strings.TrimSpace(raw))
	return strings.Join(parts, " ")
}

func normalizeDisplayText(raw string) string {
	normalized := normalizeCellText(raw)
	if normalized == "" {
		return normalized
	}

	builder := strings.Builder{}
	builder.Grow(len(normalized))
	shouldUppercaseNext := true

	for _, char := range normalized {
		switch {
		case unicode.IsSpace(char):
			builder.WriteRune(char)
			shouldUppercaseNext = true
		case isDisplayWordSeparator(char):
			builder.WriteRune(char)
			shouldUppercaseNext = true
		case shouldUppercaseNext:
			builder.WriteRune(unicode.ToUpper(char))
			shouldUppercaseNext = false
		default:
			builder.WriteRune(unicode.ToLower(char))
		}
	}

	return builder.String()
}

func normalizeLookupKey(raw string) string {
	return strings.ToLower(normalizeCellText(raw))
}

func salespersonFirstNameLookupKey(raw string) string {
	normalized := normalizeLookupKey(raw)
	if normalized == "" || normalized == normalizeLookupKey(notInformedName) {
		return ""
	}

	parts := strings.Fields(normalized)
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

func buildSalespersonFirstNameCounts(items []model.SalespersonModel) map[string]int {
	counts := make(map[string]int)
	for _, item := range items {
		firstNameKey := salespersonFirstNameLookupKey(item.Name)
		if firstNameKey == "" {
			continue
		}

		counts[firstNameKey]++
	}

	return counts
}

func buildCanonicalSalespersonFirstNameIDs(items []model.SalespersonModel) map[string]int64 {
	canonicalIDs := make(map[string]int64)
	for _, item := range items {
		nameKey := normalizeLookupKey(item.Name)
		firstNameKey := salespersonFirstNameLookupKey(item.Name)
		if firstNameKey == "" || nameKey == "" || nameKey != firstNameKey {
			continue
		}

		canonicalIDs[firstNameKey] = item.ID
	}

	return canonicalIDs
}

func isDisplayWordSeparator(char rune) bool {
	switch char {
	case '-', '/', '(', ')', '.':
		return true
	default:
		return false
	}
}

func resolveCatalogName(
	raw string,
	label string,
	existing map[string]struct{},
	defaultExists bool,
	options dto.PreviewBudgetImportOptions,
	missing map[string]struct{},
	messages *[]string,
) (string, bool, bool) {
	if isMissingValue(raw) {
		if !options.UseDefaultNotInformed {
			return "", false, false
		}
		if defaultExists {
			*messages = append(*messages, fmt.Sprintf("%s nao informado, sera usado Nao informado.", label))
			return normalizeLookupKey(notInformedName), true, false
		}
		if options.CreateMissingCatalogs {
			missing[normalizeLookupKey(notInformedName)] = struct{}{}
			*messages = append(*messages, fmt.Sprintf("%s nao informado, item Nao informado sera criado automaticamente.", label))
			return normalizeLookupKey(notInformedName), true, false
		}
		*messages = append(*messages, fmt.Sprintf("%s nao informado e item padrao Nao informado nao existe.", label))
		return "", false, true
	}

	name := normalizeDisplayText(raw)
	key := normalizeLookupKey(name)
	if _, ok := existing[key]; ok {
		return key, false, false
	}
	if options.CreateMissingCatalogs {
		missing[key] = struct{}{}
		*messages = append(*messages, fmt.Sprintf("%s sera criado automaticamente: %s.", label, name))
		return key, true, false
	}

	*messages = append(*messages, fmt.Sprintf("%s nao encontrado: %s.", label, name))
	return "", false, true
}

func resolveContactName(
	raw string,
	installerName string,
	existing map[string]struct{},
	defaultExists bool,
	options dto.PreviewBudgetImportOptions,
	missing map[string]struct{},
	messages *[]string,
) (string, bool, bool) {
	if isMissingValue(raw) {
		if !options.UseDefaultNotInformed {
			return "", false, false
		}
		if defaultExists {
			*messages = append(*messages, "Contato nao informado, sera usado Nao informado.")
			return installerName + "|" + normalizeLookupKey(notInformedName), true, false
		}
		if options.CreateMissingCatalogs {
			missing[installerName+"|"+normalizeLookupKey(notInformedName)] = struct{}{}
			*messages = append(*messages, "Contato nao informado, item Nao informado sera criado automaticamente.")
			return installerName + "|" + normalizeLookupKey(notInformedName), true, false
		}
		*messages = append(*messages, "Contato nao informado e item padrao Nao informado nao existe.")
		return "", false, true
	}

	name := normalizeDisplayText(raw)
	key := installerName + "|" + normalizeLookupKey(name)
	if _, ok := existing[key]; ok {
		return key, false, false
	}
	if options.CreateMissingCatalogs {
		missing[key] = struct{}{}
		*messages = append(*messages, fmt.Sprintf("Contato sera criado automaticamente: %s.", name))
		return key, true, false
	}

	*messages = append(*messages, fmt.Sprintf("Contato nao encontrado: %s.", name))
	return "", false, true
}

func resolveSalespersonName(
	raw string,
	existing map[string]struct{},
	defaultExists bool,
	options dto.PreviewBudgetImportOptions,
	missing map[string]struct{},
	messages *[]string,
) (string, bool, bool) {
	if isMissingValue(raw) {
		return resolveCatalogName(
			raw,
			"Vendedor",
			existing,
			defaultExists,
			options,
			missing,
			messages,
		)
	}

	name := normalizeDisplayText(raw)
	key := normalizeLookupKey(name)
	firstNameKey := salespersonFirstNameLookupKey(name)
	if firstNameKey != "" {
		if _, ok := existing[firstNameKey]; ok {
			return firstNameKey, false, false
		}
	}
	if _, ok := existing[key]; ok {
		return key, false, false
	}

	if options.CreateMissingCatalogs {
		missing[key] = struct{}{}
		*messages = append(*messages, fmt.Sprintf("Vendedor sera criado automaticamente: %s.", name))
		return key, true, false
	}

	*messages = append(*messages, fmt.Sprintf("Vendedor nao encontrado: %s.", name))
	return "", false, true
}

func ensureDefaultPriority(
	defaultExists bool,
	options dto.PreviewBudgetImportOptions,
	missing map[string]struct{},
	messages *[]string,
) (bool, bool) {
	if defaultExists {
		return false, false
	}
	if options.CreateMissingCatalogs && options.UseDefaultNotInformed {
		missing[normalizeLookupKey(notInformedName)] = struct{}{}
		*messages = append(*messages, "Prioridade operacional sera tratada como Nao informado.")
		return true, false
	}
	if options.UseDefaultNotInformed {
		*messages = append(*messages, "Prioridade operacional Nao informado nao existe.")
		return false, true
	}
	return false, false
}

func ensureDefaultCatalog(defaultExists bool, missing map[string]struct{}, options dto.PreviewBudgetImportOptions) {
	if defaultExists || !options.UseDefaultNotInformed || !options.CreateMissingCatalogs {
		return
	}
	missing[normalizeLookupKey(notInformedName)] = struct{}{}
}

func markRowError(row dto.BudgetImportPreviewRow, message string) dto.BudgetImportPreviewRow {
	row.Status = "error"
	row.Action = "error"
	row.Messages = append(row.Messages, message)
	return row
}

func resultWithError() rowPreviewResult {
	return rowPreviewResult{hasError: true}
}
