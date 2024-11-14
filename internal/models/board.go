package models

import "database/sql"

type Board struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type BoardRepository struct {
	db *sql.DB
}

func NewBoardRepository(db *sql.DB) *BoardRepository {
	return &BoardRepository{db}
}
func (br *BoardRepository) GetBoardByID(id string) (Board, error) {
	rows := br.db.QueryRow("SELECT id, name FROM boards where id = :1", id)
	var board Board
	err := rows.Scan(&board.ID, &board.Name)
	if err != nil {
		return Board{}, err
	}
	return board, nil
}
func (br *BoardRepository) GetBoardsByUserId(userId string) ([]Board, error) {
	rows, err := br.db.Query("SELECT b.id, b.name FROM boards b join users_boards ub on b.id = ub.board_id where ub.user_id = :1", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	boards := []Board{}
	for rows.Next() {
		var board Board
		err := rows.Scan(&board.ID, &board.Name)
		if err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	return boards, nil
}
