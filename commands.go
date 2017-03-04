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

const (
	up = iota
	right
	down
	left
)

var directions = [4]string{
	up:    "up",
	down:  "down",
	left:  "left",
	right: "right",
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

	board := buildBoard(data)
	myhead := getSnake(data, data.You).Coords[0]

	dir := 0

	if myhead.Y > 0 && board[myhead.Y-1][myhead.X].t != SNAKE {
		dir = up
	} else if myhead.Y < data.Height-1 && board[myhead.Y+1][myhead.X].t != SNAKE {
		dir = down
	} else if myhead.X < 0 && board[myhead.Y][myhead.X-1].t != SNAKE {
		dir = left
	} else if myhead.X < data.Width-1 && board[myhead.Y][myhead.X+1].t != SNAKE {
		dir = right
	}

	respond(res, MoveResponse{
		Move:  directions[dir],
		Taunt: &data.You,
	})
}
