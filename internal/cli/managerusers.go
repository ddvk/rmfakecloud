package cli

import (
	"flag"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage/fs"
	log "github.com/sirupsen/logrus"
)

func Handler(cfg *config.Config, args []string) bool {
	if len(args) > 1 && args[1] == "setuser" {
		userParam := flag.NewFlagSet("adduser", flag.ExitOnError)
		username := userParam.String("u", "", "username")
		pass := userParam.String("p", "", "password")

		userParam.Parse(args[2:])
		if *username != "" && *pass != "" {

			storage := fs.Storage{
				Cfg: cfg,
			}

			usr, err := storage.GetUser(*username)
			if err == nil {
				log.Info("Updateing user: ", *username)
				usr.SetPassword(*pass)
				storage.UpdateUser(usr)
			} else {
				log.Info("Creating user: ", *username)
				usr := &model.User{
					Id:    *username,
					Email: *username,
				}

				err := storage.RegisterUser(usr)
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			userParam.PrintDefaults()
		}
		return true
	}
	return false

}
