package controllers

func (controller Controller) Select(path string) ([]string, error) {
	var paths []string
	var err error
	if len(path) > 0 {
		paths, err = controller.UnfoldPath(path)
		if err != nil {
			return nil, err
		}
	}

	if len(paths) == 1 && paths[0] == path {
		err = controller.CD(paths[0])
		if err != nil {
			return nil, err
		}
	}

	return controller.SetClipBoard(paths)
}
