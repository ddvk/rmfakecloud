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
		admin := userParam.Bool("a", false, "admin role")

		userParam.Parse(args[2:])
		if *username != "" && *pass != "" {

			storage := fs.Storage{
				Cfg: cfg,
			}

			usr, err := storage.GetUser(*username)
			if err != nil {
				usr = &model.User{
					Id:    *username,
					Email: *username,
				}

			}
			
			if (usr == nil) {

				newUser, err := model.NewUser(*username, *pass)

				if err != nil {
					log.Error("error NewUser:", err)					
					return false;
				}

				newUser.SetPassword(*pass)
				newUser.IsAdmin = *admin
			
				err = storage.RegisterUser(newUser)
				if err != nil {
					log.Error("error RegisterUser:", err)					
					return false;
				}

				log.Info("Created the user")

			} else {

				usr.SetPassword(*pass)
				usr.IsAdmin = *admin
				err = storage.UpdateUser(usr)
				if err != nil {
					log.Fatal(err)
				}

				log.Info("Updated the user")
			}


		} else {
			userParam.PrintDefaults()
		}
		return true
	}
	return false

}
