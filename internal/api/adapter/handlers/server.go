package handlers

import "avito-test-applicant/internal/service"

type Server struct {
	Services *service.Services
}

func NewServer(
	services *service.Services,
) *Server {
	return &Server{
		Services: services,
	}
}
