package game

import (
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/leaderboard"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/pet"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/rewards"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/summary"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/tasks"
)

type Handler struct {
	pet         pet.Service
	tasks       tasks.Service
	rewards     rewards.Service
	leaderboard leaderboard.Service
	summary     summary.Service
}

func NewHandler(petService pet.Service, tasksService tasks.Service, rewardsService rewards.Service, leaderboardService leaderboard.Service, summaryService summary.Service) *Handler {
	return &Handler{pet: petService, tasks: tasksService, rewards: rewardsService, leaderboard: leaderboardService, summary: summaryService}
}
