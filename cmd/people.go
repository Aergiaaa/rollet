package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/Aergiaaa/rollet/internal/database"
	"github.com/gin-gonic/gin"
)

type PersonInput struct {
	Name string `json:"name" binding:"required"`
	Role string `json:"role" binding:"required"`
}

type RandomizeRequest struct {
	People    []PersonInput `json:"people" binding:"required,min=1"`
	TeamCount int           `json:"team_count" binding:"required,min=1"`
}

type RoleGroup struct {
	Role   string             `json:"role"`
	People []*database.People `json:"people"`
}

type TeamGroup struct {
	Team    int                `json:"team"`
	Members []*database.People `json:"members"`
}

type RandomizeResponse struct {
	Teams []TeamGroup `json:"teams"`
	Total int         `json:"total"`
}

func (app *app) createRandomize(c *gin.Context) {

	// Bind and validate input
	var req RandomizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Convert input to People structs
	people := make([]*database.People, len(req.People))
	for i, p := range req.People {
		people[i] = &database.People{
			Name: p.Name,
			Role: p.Role,
			Team: 0,
		}
	}

	// Group by role
	roleMap := make(map[string][]*database.People)
	for _, person := range people {
		roleMap[person.Role] = append(roleMap[person.Role], person)
	}

	// Shuffle and assign to teams
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	allPeople := make([]*database.People, 0, len(people))

	for _, peopleList := range roleMap {
		shuffled := make([]*database.People, len(peopleList))
		copy(shuffled, peopleList)

		rng.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		// round-robin assignment to teams
		for i, person := range shuffled {
			person.Team = (i % req.TeamCount) + 1
			allPeople = append(allPeople, person)
		}
	}

	// Save to database if authenticated
	user, exists := c.Get("user")
	isAuthenticated := exists && user != nil
	if isAuthenticated {
		userObj := user.(*database.User)
		err := app.models.People.Save(userObj.Id, allPeople)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save to database",
			})
			return
		}
	}

	// Group by team for response
	teamMap := make(map[int][]*database.People)
	for _, person := range allPeople {
		teamMap[person.Team] = append(teamMap[person.Team], person)
	}

	// Prepare teams for response
	teams := make([]TeamGroup, 0, len(teamMap))
	for teamNum := 1; teamNum <= req.TeamCount; teamNum++ {
		if peopleInTeam, ok := teamMap[teamNum]; ok {
			teams = append(teams, TeamGroup{
				Team:    teamNum,
				Members: peopleInTeam,
			})
		}
	}

	res := RandomizeResponse{
		Teams: teams,
		Total: len(allPeople),
	}
	c.JSON(http.StatusOK, res)
}

func (app *app) getHistory(c *gin.Context) {

	// Check authentication
	user, exists := c.Get("user")
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Retrieve saved data
	userObj := user.(*database.User)
	people, err := app.models.People.GetAllbyUserId(userObj.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve data",
		})
		return
	}

	teamMap := make(map[int][]*database.People)
	for _, person := range people {
		teamMap[person.Team] = append(teamMap[person.Team], person)
	}

	teams := make([]TeamGroup, 0, len(teamMap))
	for teamNum, peopleInTeam := range teamMap {
		teams = append(teams, TeamGroup{
			Team:    teamNum,
			Members: peopleInTeam,
		})
	}

	res := RandomizeResponse{
		Teams: teams,
		Total: len(people),
	}
	c.JSON(http.StatusOK, res)
}
