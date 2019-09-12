package types

func GenerateResourceUrl(urlPrefix string, resource Object) string {
	var prefix string
	schema := resource.GetSchema()
	if parent := resource.GetParent(); parent != nil {
		prefix = GenerateResourceUrl(urlPrefix, parent)
	} else {
		prefix = urlPrefix + schema.Version.GetUrl()
	}
	return prefix + "/" + schema.GetCollectionName() + "/" + resource.GetID()
}

func GenerateResourceCollectionUrl(urlPrefix string, resource Object) string {
	var prefix string
	schema := resource.GetSchema()
	if parent := resource.GetParent(); parent != nil {
		prefix = GenerateResourceUrl(urlPrefix, parent)
	} else {
		prefix = urlPrefix + schema.Version.GetUrl()
	}
	return prefix + "/" + resource.GetSchema().GetCollectionName()
}

func GenerateChildrenUrl(urlPrefix string, resource Object) map[string]string {
	schema := resource.GetSchema()
	prefix := GenerateResourceUrl(urlPrefix, resource)
	urls := make(map[string]string)
	for _, child := range schema.children {
		collectionName := child.GetCollectionName()
		urls[collectionName] = prefix + "/" + collectionName
	}
	return urls
}
