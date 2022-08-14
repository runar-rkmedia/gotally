package tallylogic

// These are a collection of constructed gamepositions that are intended to
// give a challenge

func NewDailyBoard() *TableBoard {
	return &TableBoard{
		rows:    5,
		columns: 5,
		cells: cellCreator(
			0, 2, 1, 0, 1,
			64, 4, 4, 1, 2,
			64, 8, 4, 1, 0,
			12, 3, 1, 0, 0,
			16, 0, 0, 0, 0,
		),
	}
}
