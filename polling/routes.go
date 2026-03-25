package polling

import (
	"polling-system/middleware"
	"polling-system/polling/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all polling routes on the given router group.
func RegisterRoutes(rg *gin.RouterGroup) {
	// Polls
	polls := rg.Group("/polls")
	{
		polls.POST("", middleware.RoleRequired("admin", "poller"), handlers.CreatePoll)
		polls.GET("", handlers.ListPolls)
		polls.GET("/:id", handlers.GetPoll)
		polls.PATCH("/:id", middleware.RoleRequired("admin", "poller"), handlers.UpdatePoll)
		polls.DELETE("/:id", middleware.RoleRequired("admin"), handlers.DeletePoll)
		polls.POST("/:id/candidates", middleware.RoleRequired("admin", "poller"), handlers.CreateCandidate)
		polls.GET("/:id/candidates", handlers.ListCandidates)
		polls.POST("/:id/candidates/bulk", middleware.RoleRequired("admin", "poller"), handlers.BulkCreateCandidates)
		polls.POST("/:id/candidates/upload", middleware.RoleRequired("admin", "poller"), handlers.UploadCandidateCSV)
		polls.POST("/:id/questions", middleware.RoleRequired("admin", "poller"), handlers.CreateQuestion)
		polls.GET("/:id/questions", handlers.ListQuestions)
	}

	// Candidates (top-level)
	candidates := rg.Group("/candidates")
	{
		candidates.GET("", handlers.QueryCandidates)
		candidates.GET("/:id", handlers.GetCandidate)
		candidates.PATCH("/:id", middleware.RoleRequired("admin", "poller"), handlers.UpdateCandidate)
		candidates.DELETE("/:id", middleware.RoleRequired("admin", "poller"), handlers.DeleteCandidate)
		candidates.GET("/:id/answers", handlers.ListAnswersByCandidate)
	}

	// Questions (top-level)
	questions := rg.Group("/questions")
	{
		questions.GET("/:id", handlers.GetQuestion)
		questions.PATCH("/:id", middleware.RoleRequired("admin", "poller"), handlers.UpdateQuestion)
		questions.DELETE("/:id", middleware.RoleRequired("admin", "poller"), handlers.DeleteQuestion)
		questions.POST("/:id/answers", middleware.RoleRequired("admin", "poller"), handlers.CreateAnswer)
		questions.GET("/:id/answers", handlers.ListAnswersByQuestion)
	}

	// Answers (top-level)
	answers := rg.Group("/answers")
	{
		answers.GET("/:id", handlers.GetAnswer)
		answers.DELETE("/:id", middleware.RoleRequired("admin"), handlers.DeleteAnswer)
	}
}
