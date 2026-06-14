package dto

type PreviewBudgetImportOptions struct {
	DuplicateStrategy     string `json:"duplicate_strategy"`
	CreateMissingCatalogs bool   `json:"create_missing_catalogs"`
	UseDefaultNotInformed bool   `json:"use_default_not_informed"`
}

type ExecuteBudgetImportRequest struct {
	PreviewID string `json:"preview_id" validate:"required"`
}

type BudgetImportPreviewMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BudgetImportPreviewSummary struct {
	RowsRead         int `json:"rows_read"`
	RowsValid        int `json:"rows_valid"`
	RowsWithWarning  int `json:"rows_with_warning"`
	RowsWithError    int `json:"rows_with_error"`
	RowsEmptyIgnored int `json:"rows_empty_ignored"`
	NewBudgets       int `json:"new_budgets"`
	ExistingBudgets  int `json:"existing_budgets"`
}

type BudgetImportCatalogActions struct {
	BudgetStatusesToCreate int `json:"budget_statuses_to_create"`
	LossReasonsToCreate    int `json:"loss_reasons_to_create"`
	InstallersToCreate     int `json:"installers_to_create"`
	ProjectsToCreate       int `json:"projects_to_create"`
	ProjectTypesToCreate   int `json:"project_types_to_create"`
	SalespeopleToCreate    int `json:"salespeople_to_create"`
	ContactsToCreate       int `json:"contacts_to_create"`
	PrioritiesToCreate     int `json:"priorities_to_create"`
}

type BudgetImportPreviewRow struct {
	RowNumber    int      `json:"row_number"`
	BudgetNumber string   `json:"budget_number"`
	Status       string   `json:"status"`
	Action       string   `json:"action"`
	Messages     []string `json:"messages"`
}

type PreviewBudgetImportResponse struct {
	PreviewID      string                       `json:"preview_id"`
	FileName       string                       `json:"file_name"`
	SheetName      string                       `json:"sheet_name"`
	HeaderRow      int                          `json:"header_row"`
	Options        PreviewBudgetImportOptions   `json:"options"`
	Summary        BudgetImportPreviewSummary   `json:"summary"`
	CatalogActions BudgetImportCatalogActions   `json:"catalog_actions"`
	Warnings       []BudgetImportPreviewMessage `json:"warnings"`
	Errors         []BudgetImportPreviewMessage `json:"errors"`
	SampleRows     []BudgetImportPreviewRow     `json:"sample_rows"`
	InconsistencyRows []BudgetImportPreviewRow  `json:"inconsistency_rows"`
}

type ExecuteBudgetImportSummary struct {
	RowsProcessed   int `json:"rows_processed"`
	BudgetsCreated  int `json:"budgets_created"`
	BudgetsUpdated  int `json:"budgets_updated"`
	BudgetsIgnored  int `json:"budgets_ignored"`
	RowsFailed      int `json:"rows_failed"`
	CatalogsCreated int `json:"catalogs_created"`
}

type ExecuteBudgetImportResult struct {
	Message string `json:"message"`
}

type ExecuteBudgetImportResponse struct {
	ImportID   string                     `json:"import_id"`
	PreviewID  string                     `json:"preview_id"`
	Status     string                     `json:"status"`
	StartedAt  string                     `json:"started_at"`
	FinishedAt string                     `json:"finished_at"`
	Summary    ExecuteBudgetImportSummary `json:"summary"`
	Result     ExecuteBudgetImportResult  `json:"result"`
}
