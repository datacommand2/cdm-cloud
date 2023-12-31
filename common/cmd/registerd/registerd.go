package main

import (
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/google/uuid"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/auth"
	"github.com/micro/go-micro/v2/config/cmd"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"

	"fmt"
	"os"
	"strings"

	_ "github.com/go-micro/plugins/v2/registry/kubernetes"
)

// OsExiter is the function used when the app exits. If not set defaults to os.Exit.
var OsExiter = os.Exit

// Service 는 service registry 에 서비스를 등록하기 위한 구조체이다.
type Service struct {
	Name          string
	Version       string
	AdvertisePort string
	Metadata      map[string]string
}

// Register 는 service registry 에 서비스를 등록하는 함수이다.
func (s Service) Register() error {
	if registryName := s.Metadata["MICRO_REGISTRY"]; registryName != "" {
		fc, ok := cmd.DefaultRegistries[registryName]
		if !ok {
			return errors.New("invalid registry name")
		}

		registry.DefaultRegistry = fc()
	}

	srv := server.NewServer(
		server.Id(uuid.New().String()),
		server.Advertise(fmt.Sprintf(":%s", s.AdvertisePort)),
		server.Metadata(s.Metadata),
	)

	return micro.NewService(
		micro.Auth(auth.NewAuth()),
		micro.Server(srv),
		micro.Name(s.Name),
		micro.Version(s.Version),
	).Run()
}

func getEnv(key string, required bool) string {
	v, ok := os.LookupEnv(key)
	if !ok && required {
		logger.Errorf("'%s' environment is required", key)
		OsExiter(1)
	}

	return v
}

func parseMetadata(s string) map[string]string {
	md := make(map[string]string)

	if s == "" {
		return md
	}

	for _, item := range strings.Split(s, ",") {
		v := strings.Split(item, "=")
		if len(v) == 2 {
			md[v[0]] = v[1]
		}
	}

	return md
}

func main() {
	md := parseMetadata(getEnv("CDM_SERVICE_METADATA", false))
	md["CDM_SOLUTION_NAME"] = getEnv("CDM_SOLUTION_NAME", true)
	md["CDM_SERVICE_DESCRIPTION"] = getEnv("CDM_SERVICE_DESCRIPTION", true)
	md["MICRO_REGISTRY"] = getEnv("MICRO_REGISTRY", false)

	s := Service{
		Name:          getEnv("CDM_SERVICE_NAME", true),
		Version:       getEnv("CDM_SERVICE_VERSION", true),
		AdvertisePort: getEnv("CDM_SERVICE_ADVERTISE_PORT", true),
		Metadata:      md,
	}

	if err := s.Register(); err != nil {
		logger.Errorf("Could not run micro service. cause: %v", err)
		OsExiter(1)
	}
}
