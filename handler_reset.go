package main

import (
	"fmt"
	"context"
)

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't delete users: %w", err)
	}
	err = s.db.ResetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't delete feeds: %w", err)
	}
	fmt.Println("database successfully reset")
	return nil
}