package tron

// 将 client 分类的组信息
type ClientGroup struct {
	Gid string
	Key string
}

func NewClientsGroup(gid, key string) *ClientGroup {
	g := &ClientGroup{
		Gid: gid,
		Key: key,
	}
	return g
}
