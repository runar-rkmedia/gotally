package api

import (
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	logic "github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

func toGameSwipeDirection(dir model.SwipeDirection) logic.SwipeDirection {
	switch dir {
	case model.SwipeDirection_SWIPE_DIRECTION_UP:
		return logic.SwipeDirectionUp
	case model.SwipeDirection_SWIPE_DIRECTION_RIGHT:
		return logic.SwipeDirectionRight
	case model.SwipeDirection_SWIPE_DIRECTION_DOWN:
		return logic.SwipeDirectionDown
	case model.SwipeDirection_SWIPE_DIRECTION_LEFT:
		return logic.SwipeDirectionLeft
	}
	return ""
}
func toModalDirection(dir logic.SwipeDirection) model.SwipeDirection {
	switch dir {
	case logic.SwipeDirectionUp:
		return model.SwipeDirection_SWIPE_DIRECTION_UP
	case logic.SwipeDirectionRight:
		return model.SwipeDirection_SWIPE_DIRECTION_RIGHT
	case logic.SwipeDirectionDown:
		return model.SwipeDirection_SWIPE_DIRECTION_DOWN
	case logic.SwipeDirectionLeft:
		return model.SwipeDirection_SWIPE_DIRECTION_LEFT
	}
	return model.SwipeDirection_SWIPE_DIRECTION_UNSPECIFIED
}

func toModalBoard(game *logic.Game) *model.Board {
	return &model.Board{
		Id:      game.BoardID(),
		Cells:   toModalCells(game.Cells()),
		Columns: int32(game.Rules.SizeX),
		Rows:    int32(game.Rules.SizeX),
		Name:    game.Name,
	}
}

func toModalCells(cells []cell.Cell) []*model.Cell {
	c := make([]*model.Cell, len(cells))
	for i := 0; i < len(cells); i++ {
		base, twopow := cells[i].Raw()
		c[i] = &model.Cell{
			Base:   base,
			Twopow: twopow,
		}

	}
	return c
}
