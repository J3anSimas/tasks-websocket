package models

import "database/sql"

type Card struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db}
}
func (cr *CardRepository) GetCardById(id string) (Card, error) {
	rows := cr.db.QueryRow("SELECT id, id, password FROM cards where id = :1", id)

	var card Card
	err := rows.Scan(&card.ID, &card.Name, &card.Status)

	if err != nil {
		return Card{}, err
	}
	return card, nil
}

func (cr *CardRepository) GetCardsByBoardId(boardId string) ([]Card, error) {
	rows, err := cr.db.Query("SELECT ID, NAME, STATUS FROM CARDS WHERE boardId = :1", boardId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []Card
	for rows.Next() {
		var card Card
		err := rows.Scan(&card.ID, &card.Name, &card.Status)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}
func (cr *CardRepository) UpdateCardStatus(cardId string, status string) error {
	_, err := cr.db.Exec("UPDATE cards SET status = :1 WHERE id = :2", status, cardId)
	if err != nil {
		return err
	}
	return nil
}
