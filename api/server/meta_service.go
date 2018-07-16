package server

const MetadataApi = "meta"

// RPCMetaService gives meta information about the server.
type RPCMetaService struct {
	server *Server
}

func (s *RPCMetaService) List() map[string]string {
	modules := make(map[string]string)
	for name := range s.server.services {
		modules[name] = "1.0"
	}
	return modules
}
