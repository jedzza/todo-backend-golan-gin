package routes

import (
	"todoappgo/Controllers"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.Engine) {
	router.POST("/user", controllers.CreateUser())
	router.GET("/user/:email", controllers.GetAUser())
	router.PUT("/user/:email", controllers.EditAUser())
	router.DELETE("/user/:email", controllers.DeleteAUser())
	router.GET("/tasks/:email", controllers.ReturnTasks())
	// router.GET("/users", controllers.GetAllUsers())
}

//func UserRoute() *gin.Engine {
//	router := gin.Default()
//	api := router.Group("/api")
//	{
//		router.POST("/signup", controllers.CreateUser())
//		router.POST("/login", controllers.Login())

//		secured := api.Group("/secured").Use(middleware.auth())
//		{
//			router.GET("/user/:email", controllers.GetAUser())
//			router.PUT("/user/:email", controllers.EditAUser())
//			router.DELETE("/user/:email", controllers.DeleteAUser())
//		}
//	}
//}
