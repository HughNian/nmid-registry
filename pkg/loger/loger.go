package loger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Loger = logrus.New()

func Init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.WarnLevel)
}
