package global

import (
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/ppxb/oreo-admin-go/pkg/config"
)

var (
	Mode        string
	RuntimeRoot string
	Conf        Configuration
	ConfBox     config.ConfBox
	Tracer      *trace.TracerProvider
	Mysql       *gorm.DB
)
