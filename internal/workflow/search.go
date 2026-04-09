package workflow

import (
	"fmt"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
)

// ShowInteractiveSelectionFlow 承接实时搜索后的多页面 TUI 流：
// 选番、选集、选线路、播放，以及播放后的回跳。
func ShowInteractiveSelectionFlow(application *app.App, animes []source.Anime, sourceName string) error {
	if len(animes) == 0 {
		return nil
	}

	if !tui.SupportsInteractiveTUI() {
		return fmt.Errorf("当前终端不支持交互式 TUI")
	}

	return showInteractiveSelectionFlow(application, animes, sourceName)
}
