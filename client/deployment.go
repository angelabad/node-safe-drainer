package client

type Deployment struct {
	Namespace string
	Name      string
}

type Deployments []Deployment

func (d *Deployments) deduplicate() {
	keys := make(map[Deployment]bool)
	list := Deployments{}
	for _, entry := range *d {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	*d = list
}
