package report

type RecordRepo interface {
	Store(Record) error
}
