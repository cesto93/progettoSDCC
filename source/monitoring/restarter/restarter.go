package restarter

type Restarter interface {
	Restart(instanceId string) (bool, error)
}