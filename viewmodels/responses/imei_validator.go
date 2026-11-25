package responses

// ImeiBatch is one uploaded batch ("lote") of the IMEI Validator.
type ImeiBatch struct {
	ID               string  `json:"id"`
	Code             string  `json:"code"`
	FileName         string  `json:"file_name"`
	CreatedByName    string  `json:"created_by_name"`
	CreatedAt        string  `json:"created_at"`
	Total            int     `json:"total"`
	Status           string  `json:"status"`
	RegisteredAt     *string `json:"registered_at"`
	RegisteredByName string  `json:"registered_by_name"`
}

type ImeiBatchUploadSummary struct {
	TotalRows         int `json:"total_rows"`
	New               int `json:"new"`
	AlreadyReady      int `json:"already_ready"`
	AlreadyRegistered int `json:"already_registered"`
	DuplicatedInFile  int `json:"duplicated_in_file"`
	Invalid           int `json:"invalid"`
	// DeviceIncomplete is the part of Invalid with reason device_incomplete:
	// valid new IMEIs not saved because their device could not be saved whole.
	DeviceIncomplete int `json:"device_incomplete"`
}

type ImeiAlreadyReady struct {
	IMEI      string `json:"imei"`
	BatchID   string `json:"batch_id"`
	BatchCode string `json:"batch_code"`
	CreatedAt string `json:"created_at"`
}

type ImeiAlreadyRegistered struct {
	IMEI         string  `json:"imei"`
	BatchID      string  `json:"batch_id"`
	BatchCode    string  `json:"batch_code"`
	RegisteredAt *string `json:"registered_at"`
}

type ImeiDuplicatedInFile struct {
	IMEI string `json:"imei"`
	Rows []int  `json:"rows"`
}

// ImeiInvalidRow is a row not saved. reason: format | luhn | serial |
// quantity | sn_count | device_incomplete. detail explains sn_count and
// device_incomplete.
type ImeiInvalidRow struct {
	Row    int    `json:"row"`
	Value  string `json:"value"`
	Serial string `json:"serial"`
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

// ImeiBatchUploadResult is the POST /batches response. Batch is null when
// the file had no new IMEI (no batch is created then).
type ImeiBatchUploadResult struct {
	Batch             *ImeiBatch              `json:"batch"`
	Summary           ImeiBatchUploadSummary  `json:"summary"`
	AlreadyReady      []ImeiAlreadyReady      `json:"already_ready"`
	AlreadyRegistered []ImeiAlreadyRegistered `json:"already_registered"`
	DuplicatedInFile  []ImeiDuplicatedInFile  `json:"duplicated_in_file"`
	Invalid           []ImeiInvalidRow        `json:"invalid"`
}

type ImeiBatchPage struct {
	Items []ImeiBatch `json:"items"`
	Total int64       `json:"total"`
}

type ImeiBatchRecord struct {
	IMEI         string  `json:"imei"`
	Serial       string  `json:"serial"`
	ImeiCount    int     `json:"imei_count"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
	RegisteredAt *string `json:"registered_at"`
}

type ImeiBatchRecordPage struct {
	Items []ImeiBatchRecord `json:"items"`
	Total int64             `json:"total"`
}

type ImeiBatchRef struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

// ImeiValidatorLookup is GET /imeis/:imei. Only imei and found are set when
// the IMEI is not stored.
type ImeiValidatorLookup struct {
	Found        bool          `json:"found"`
	IMEI         string        `json:"imei"`
	Status       string        `json:"status"`
	Batch        *ImeiBatchRef `json:"batch"`
	Serial       string        `json:"serial"`
	ImeiCount    int           `json:"imei_count"`
	CreatedAt    string        `json:"created_at"`
	RegisteredAt *string       `json:"registered_at"`
}

// ImeiDeleteResult: deleting an IMEI removes its whole device (every IMEI of
// the same S/N in the batch), listed in deleted_imeis, so the OABI file never
// has an incomplete device.
type ImeiDeleteResult struct {
	BatchDeleted bool       `json:"batch_deleted"`
	Batch        *ImeiBatch `json:"batch"`
	DeletedIMEIs []string   `json:"deleted_imeis"`
}
