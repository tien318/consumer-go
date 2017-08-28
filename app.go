package consumer

// App lorem
type App struct {
	ID      int
	AppCode string
	AppName string
}

// AppService lorem
type AppService interface {
	GetByAppCode(appCode string) (*App, error)
}
