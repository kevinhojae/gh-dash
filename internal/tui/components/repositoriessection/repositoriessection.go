package repositoriessection

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/repository"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

const SectionType = "repositories"

type Model struct {
	section.BaseModel
	Repositories []data.RepositoryData
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.RepositoriesSectionConfig,
	lastUpdated time.Time,
	createdAt time.Time,
) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      cfg.ToSectionConfig(),
			Type:        SectionType,
			Columns:     GetSectionColumns(cfg, ctx),
			Singular:    m.GetItemSingularForm(),
			Plural:      m.GetItemPluralForm(),
			LastUpdated: lastUpdated,
			CreatedAt:   createdAt,
		},
	)
	m.Repositories = []data.RepositoryData{}

	return m
}

func (m *Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:

		if m.IsSearchFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.SearchBar.SetValue(m.SearchValue)
				blinkCmd := m.SetIsSearching(false)
				return m, blinkCmd

			case tea.KeyEnter:
				m.SearchValue = m.SearchBar.Value()
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

			break
		}

		switch {
		case key.Matches(msg, keys.RepositoryKeys.ToggleSmartFiltering):
			if !m.HasRepoNameInConfiguredFilter() {
				m.IsFilteredByCurrentRemote = !m.IsFilteredByCurrentRemote
			}
			searchValue := m.GetSearchValue()
			if m.SearchValue != searchValue {
				m.SearchValue = searchValue
				m.SearchBar.SetValue(searchValue)
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}
		}

	case SectionRepositoriesFetchedMsg:
		if m.LastFetchTaskId == msg.TaskId {
			if m.PageInfo != nil {
				m.Repositories = append(m.Repositories, msg.Repositories...)
			} else {
				m.Repositories = msg.Repositories
			}
			m.TotalCount = msg.TotalCount
			m.SetIsLoading(false)
			m.PageInfo = &msg.PageInfo
			m.Table.SetRows(m.BuildRows())
			m.UpdateLastUpdated(time.Now())
			m.UpdateTotalItemsCount(m.TotalCount)
		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.SearchBar = search

	table, tableCmd := m.Table.Update(msg)
	m.Table = table

	return m, tea.Batch(cmd, searchCmd, tableCmd)
}

func GetSectionColumns(
	cfg config.RepositoriesSectionConfig,
	ctx *context.ProgramContext,
) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Repositories
	sLayout := cfg.Layout

	updatedAtLayout := config.MergeColumnConfigs(
		dLayout.UpdatedAt,
		sLayout.UpdatedAt,
	)
	createdAtLayout := config.MergeColumnConfigs(
		dLayout.CreatedAt,
		sLayout.CreatedAt,
	)
	nameLayout := config.MergeColumnConfigs(dLayout.Name, sLayout.Name)
	descriptionLayout := config.MergeColumnConfigs(dLayout.Description, sLayout.Description)
	starsLayout := config.MergeColumnConfigs(dLayout.Stars, sLayout.Stars)
	languageLayout := config.MergeColumnConfigs(dLayout.Language, sLayout.Language)

	return []table.Column{
		{
			Title:  "",
			Width:  utils.IntPtr(3),
			Hidden: utils.BoolPtr(false),
		},
		{
			Title:  "Name",
			Width:  nameLayout.Width,
			Hidden: nameLayout.Hidden,
		},
		{
			Title:  "Description",
			Grow:   utils.BoolPtr(true),
			Hidden: descriptionLayout.Hidden,
		},
		{
			Title:  "⭐",
			Width:  starsLayout.Width,
			Hidden: starsLayout.Hidden,
		},
		{
			Title:  "Language",
			Width:  languageLayout.Width,
			Hidden: languageLayout.Hidden,
		},
		{
			Title:  "󱦻",
			Width:  updatedAtLayout.Width,
			Hidden: updatedAtLayout.Hidden,
		},
		{
			Title:  "󱡢",
			Width:  createdAtLayout.Width,
			Hidden: createdAtLayout.Hidden,
		},
	}
}

func (m Model) BuildRows() []table.Row {
	var rows []table.Row
	for _, currRepo := range m.Repositories {
		// GitHub API already handles filtering based on search query
		// No need for client-side filtering
		repoModel := repository.Repository{Ctx: m.Ctx, Data: currRepo}
		rows = append(rows, repoModel.ToTableRow())
	}

	if rows == nil {
		rows = []table.Row{}
	}

	return rows
}

func (m *Model) NumRows() int {
	return len(m.Repositories)
}

func (m *Model) GetCurrRow() data.RowData {
	if len(m.Repositories) == 0 {
		return nil
	}
	repo := m.Repositories[m.Table.GetCurrItem()]
	return &repo
}

func (m *Model) FetchNextPageSectionRows() []tea.Cmd {
	if m == nil {
		return nil
	}

	if m.PageInfo != nil && !m.PageInfo.HasNextPage {
		return nil
	}

	var cmds []tea.Cmd

	startCursor := time.Now().String()
	if m.PageInfo != nil {
		startCursor = m.PageInfo.StartCursor
	}
	taskId := fmt.Sprintf("fetching_repositories_%d_%s", m.Id, startCursor)
	m.LastFetchTaskId = taskId
	task := context.Task{
		Id:        taskId,
		StartText: fmt.Sprintf(`Fetching repositories for "%s"`, m.Config.Title),
		FinishedText: fmt.Sprintf(
			`Repositories for "%s" have been fetched`,
			m.Config.Title,
		),
		State: context.TaskStart,
		Error: nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	fetchCmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.RepositoriesLimit
		}
		res, err := data.FetchRepositories(m.GetFilters(), *limit, m.PageInfo)
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      taskId,
				Err:         err,
			}
		}

		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg: SectionRepositoriesFetchedMsg{
				Repositories: res.Repositories,
				TotalCount:   res.TotalCount,
				PageInfo:     res.PageInfo,
				TaskId:       taskId,
			},
		}
	}
	cmds = append(cmds, fetchCmd)

	return cmds
}

func (m *Model) UpdateLastUpdated(t time.Time) {
	m.Table.UpdateLastUpdated(t)
}

func (m *Model) ResetRows() {
	m.Repositories = nil
	m.BaseModel.ResetRows()
}

func FetchAllSections(
	ctx *context.ProgramContext,
) (sections []section.Section, fetchAllCmd tea.Cmd) {
	sectionConfigs := ctx.Config.RepositoriesSections
	fetchRepositoriesCmds := make([]tea.Cmd, 0, len(sectionConfigs))
	sections = make([]section.Section, 0, len(sectionConfigs))
	for i, sectionConfig := range sectionConfigs {
		sectionModel := NewModel(
			i+1,
			ctx,
			sectionConfig,
			time.Now(),
			time.Now(),
		) // 0 is the search section
		sections = append(sections, &sectionModel)
		fetchRepositoriesCmds = append(
			fetchRepositoriesCmds,
			sectionModel.FetchNextPageSectionRows()...)
	}
	return sections, tea.Batch(fetchRepositoriesCmds...)
}

type SectionRepositoriesFetchedMsg struct {
	Repositories []data.RepositoryData
	TotalCount   int
	PageInfo     data.PageInfo
	TaskId       string
}

func (m Model) GetItemSingularForm() string {
	return "Repository"
}

func (m Model) GetItemPluralForm() string {
	return "Repositories"
}

func (m *Model) GetFilters() string {
	filters := m.Config.Filters
	searchValue := m.GetSearchValue()

	// Remove "is:repositories" prefix as GitHub API doesn't support it
	searchValue = strings.TrimPrefix(searchValue, "is:repositories ")
	searchValue = strings.TrimSpace(searchValue)

	if searchValue == "" {
		return filters
	}

	if filters == "" {
		return searchValue
	}

	return fmt.Sprintf("%s %s", filters, searchValue)
}

func (m Model) GetTotalCount() int {
	return m.TotalCount
}

func (m *Model) GetIsLoading() bool {
	return m.IsLoading
}

func (m *Model) SetIsLoading(val bool) {
	m.IsLoading = val
	m.Table.SetIsLoading(val)
}

func (m Model) GetPagerContent() string {
	pagerContent := ""
	if m.TotalCount > 0 {
		pagerContent = fmt.Sprintf(
			"%v %v • %v %v/%v • Fetched %v",
			constants.WaitingIcon,
			m.LastUpdated().Format("01/02 15:04:05"),
			m.SingularForm,
			m.Table.GetCurrItem()+1,
			m.TotalCount,
			len(m.Table.Rows),
		)
	}
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Render(pagerContent)
	return pager
}
