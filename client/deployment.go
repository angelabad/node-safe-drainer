package client

// Deployment stores deployment namespace and name
type Deployment struct {
	Namespace string
	Name      string
}

// Deployments is an slice of deployments
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
