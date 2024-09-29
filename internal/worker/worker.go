package worker

type Worker interface {
	Init() error
	Start()
}
