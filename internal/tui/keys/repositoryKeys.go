package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

type RepositoryKeyMap struct {
	ToggleSmartFiltering key.Binding
	ViewPRs              key.Binding
}

var RepositoryKeys = RepositoryKeyMap{
	ToggleSmartFiltering: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle smart filtering"),
	),
	ViewPRs: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch to PRs"),
	),
}

func RepositoryFullHelp() []key.Binding {
	return []key.Binding{
		RepositoryKeys.ToggleSmartFiltering,
		RepositoryKeys.ViewPRs,
	}
}

func rebindRepositoryKeys(keys []config.Keybinding) error {
	CustomRepositoryBindings = []key.Binding{}

	for _, repoKey := range keys {
		if repoKey.Builtin == "" {
			// Handle custom commands
			if repoKey.Command != "" {
				name := repoKey.Name
				if repoKey.Name == "" {
					name = config.TruncateCommand(repoKey.Command)
				}

				customBinding := key.NewBinding(
					key.WithKeys(repoKey.Key),
					key.WithHelp(repoKey.Key, name),
				)

				CustomRepositoryBindings = append(CustomRepositoryBindings, customBinding)
			}
			continue
		}

		log.Debug("Rebinding repository key", "builtin", repoKey.Builtin, "key", repoKey.Key)

		var key *key.Binding

		switch repoKey.Builtin {
		case "viewPrs":
			key = &RepositoryKeys.ViewPRs
		default:
			return fmt.Errorf("unknown built-in repository key: '%s'", repoKey.Builtin)
		}

		key.SetKeys(repoKey.Key)

		helpDesc := key.Help().Desc
		if repoKey.Name != "" {
			helpDesc = repoKey.Name
		}
		key.SetHelp(repoKey.Key, helpDesc)
	}

	return nil
}
