test-watch:
	fd | entr -r gotest ./...
interfaces:
	ifacemaker -f tallylogic/board.go -s TableBoard -i BoardController -p tallylogic -o tallylogic/board_controller.go
