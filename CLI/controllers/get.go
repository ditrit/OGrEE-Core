package controllers

var Get GetController = &GetControllerImpl{}

type GetController interface {
	GetObject(path string) (map[string]any, error)
}

type GetControllerImpl struct{}

func (controller GetControllerImpl) GetObject(path string) (map[string]any, error) {
	return GetObjectWithChildren(path, 0)
}
