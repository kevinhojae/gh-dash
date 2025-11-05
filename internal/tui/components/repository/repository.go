package repository

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type Repository struct {
	Ctx  *context.ProgramContext
	Data data.RepositoryData
}

func (repo *Repository) ToTableRow() table.Row {
	return table.Row{
		repo.renderVisibility(),
		repo.renderName(),
		repo.renderDescription(),
		repo.renderStars(),
		repo.renderLanguage(),
		repo.renderUpdateAt(),
		repo.renderCreatedAt(),
	}
}

func (repo *Repository) getTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(repo.Ctx.Theme.PrimaryText)
}

func (repo *Repository) renderVisibility() string {
	if repo.Data.IsPrivate {
		return lipgloss.NewStyle().Foreground(repo.Ctx.Theme.WarningText).Render("🔒")
	}
	return repo.getTextStyle().Render("📖")
}

func (repo *Repository) renderName() string {
	return lipgloss.NewStyle().
		Foreground(repo.Ctx.Theme.PrimaryText).
		Bold(true).
		Render(repo.Data.NameWithOwner)
}

func (repo *Repository) renderDescription() string {
	description := repo.Data.Description
	if description == "" {
		description = repo.Data.Name
	}
	return repo.getTextStyle().Render(description)
}

func (repo *Repository) renderStars() string {
	stars := utils.ShortNumber(repo.Data.StargazerCount)
	return repo.getTextStyle().Render(fmt.Sprintf("⭐ %s", stars))
}

func (repo *Repository) renderLanguage() string {
	if repo.Data.PrimaryLanguage.Name == "" {
		return repo.getTextStyle().Render("-")
	}

	color := repo.Data.PrimaryLanguage.Color
	if color == "" {
		return lipgloss.NewStyle().
			Foreground(repo.Ctx.Theme.SecondaryText).
			Render(fmt.Sprintf("● %s", repo.Data.PrimaryLanguage.Name))
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(fmt.Sprintf("● %s", repo.Data.PrimaryLanguage.Name))
}

func (repo *Repository) renderUpdateAt() string {
	timeFormat := repo.Ctx.Config.Defaults.DateFormat

	updatedAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		updatedAtOutput = utils.TimeElapsed(repo.Data.UpdatedAt)
	} else {
		updatedAtOutput = repo.Data.UpdatedAt.Format(timeFormat)
	}

	return repo.getTextStyle().Render(updatedAtOutput)
}

func (repo *Repository) renderCreatedAt() string {
	timeFormat := repo.Ctx.Config.Defaults.DateFormat

	createdAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		createdAtOutput = utils.TimeElapsed(repo.Data.CreatedAt)
	} else {
		createdAtOutput = repo.Data.CreatedAt.Format(timeFormat)
	}

	return repo.getTextStyle().Render(createdAtOutput)
}
