package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CellType uint8

const (
	EMPTY CellType = iota
	SNAKE
	FOOD
)

type Cell struct {
	t     CellType
	snake string
	pos   int
}

type Dir uint8

const (
	UP Dir = iota
	RIGHT
	DOWN
	LEFT
	num_dirs
)

var directions = [4]string{
	UP:    "up",
	DOWN:  "down",
	LEFT:  "left",
	RIGHT: "right",
}

func buildBoard(req *MoveRequest) (board [][]Cell) {
	board = make([][]Cell, req.Width)
	for i := range board {
		board[i] = make([]Cell, req.Height)
	}

	for _, food := range req.Food {
		board[food.Y][food.X].t = FOOD
	}

	for _, snake := range req.Snakes {
		for i, body := range snake.Coords {
			board[body.Y][body.X] = Cell{t: SNAKE, snake: snake.Id, pos: i}
		}
	}
	return board
}

func getSnake(req *MoveRequest, id string) Snake {
	for _, snake := range req.Snakes {
		if snake.Id == id {
			return snake
		}
	}
	return Snake{}
}

func safeMove(req *MoveRequest, dir Dir) bool {
	board := buildBoard(req)
	myhead := getSnake(req, req.You).Coords[0]

	if dir == UP && myhead.Y > 0 && board[myhead.Y-1][myhead.X].t != SNAKE {
		return true
	} else if dir == DOWN && myhead.Y < req.Height-1 && board[myhead.Y+1][myhead.X].t != SNAKE {
		return true
	} else if dir == LEFT && myhead.X > 0 && board[myhead.Y][myhead.X-1].t != SNAKE {
		return true
	} else if dir == RIGHT && myhead.X < req.Width-1 && board[myhead.Y][myhead.X+1].t != SNAKE {
		return true
	}
	return false
}

func respond(res http.ResponseWriter, obj interface{}) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(obj)
}

func handleStart(res http.ResponseWriter, req *http.Request) {
	data, err := NewGameStartRequest(req)
	if err != nil {
		respond(res, GameStartResponse{
			Taunt:   toStringPointer("battlesnake-go!"),
			Color:   "#00FF00",
			Name:    fmt.Sprintf("%v (%vx%v)", data.GameId, data.Width, data.Height),
			HeadUrl: toStringPointer(fmt.Sprintf("%v://%v/static/head.png")),
		})
	}

	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	respond(res, GameStartResponse{
		Taunt:   toStringPointer("battlesnake-go!"),
		Color:   "#00FF00",
		Name:    fmt.Sprintf("%v (%vx%v)", data.GameId, data.Width, data.Height),
		HeadUrl: toStringPointer(fmt.Sprintf("%v://%v/static/head.png", scheme, req.Host)),
	})
}

func handleMove(res http.ResponseWriter, req *http.Request) {
	data, err := NewMoveRequest(req)
	if err != nil {
		respond(res, MoveResponse{
			Move:  "up",
			Taunt: toStringPointer("can't parse this!"),
		})
		return
	}

	var dir Dir
	for dir = UP; dir < num_dirs; dir++ {
		if safeMove(data, dir) {
			break
		}
	}

	respond(res, MoveResponse{
		Move:  directions[dir],
		Taunt: &data.You,
	})
}
