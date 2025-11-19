package tui

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) openBrowser() tea.Cmd {
	taskId := fmt.Sprintf("open_browser_%d", time.Now().Unix())
	task := context.Task{
		Id:           taskId,
		StartText:    "Opening in browser",
		FinishedText: "Opened in browser",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.ctx.StartTask(task)
	openCmd := func() tea.Msg {
		b := browser.New("", os.Stdout, os.Stdin)
		currRow := m.getCurrRowData()
		if currRow == nil || reflect.ValueOf(currRow).IsNil() {
			return constants.TaskFinishedMsg{
				TaskId: taskId,
				Err:    errors.New("current selection doesn't have a URL"),
			}
		}
		url := currRow.GetUrl()

		// In repositories view we want to open the Pull Requests page for the
		// selected repository rather than the repository root.
		if m.ctx.View == config.RepositoriesView && url != "" {
			url = strings.TrimRight(url, "/")
			url = fmt.Sprintf("%s/pulls", url)
		}

		err := b.Browse(url)
		return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
	}
	return tea.Batch(startCmd, openCmd)
}
