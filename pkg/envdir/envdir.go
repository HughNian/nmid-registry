package envdir

import (
	"nmid-registry/pkg/option"
	"nmid-registry/pkg/utils"
)

func InitEnvDir(opt *option.Options) error {
	err := utils.MkdirAll(opt.AbsHomeDir)
	if err != nil {
		return err
	}

	err = utils.MkdirAll(opt.AbsDataDir)
	if err != nil {
		return err
	}

	if opt.AbsWALDir != "" {
		err = utils.MkdirAll(opt.AbsWALDir)
		if err != nil {
			return err
		}
	}

	err = utils.MkdirAll(opt.AbsLogDir)
	if err != nil {
		return err
	}

	err = utils.MkdirAll(opt.AbsMemberDir)
	if err != nil {
		return err
	}

	return nil
}

func CleanEnvDir(opt *option.Options) {
	utils.RemoveAll(opt.AbsDataDir)
	if opt.AbsWALDir != "" {
		utils.RemoveAll(opt.AbsWALDir)
	}
	utils.RemoveAll(opt.AbsMemberDir)
	utils.RemoveAll(opt.AbsLogDir)
	utils.RemoveAll(opt.AbsHomeDir)
}
