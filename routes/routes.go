package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"Shamas/nutrisun/controllers"
	"Shamas/nutrisun/middleware"
	"Shamas/nutrisun/server"
)

func SetupRouter(masterDB *gorm.DB) *gin.Engine {
	r := server.InitializeServer()
	SetupAPIRoutes(r, masterDB)

	public := r.Group("/api")
	{
		public.POST("/login", middleware.StrictRateLimiterMiddleware(), func(c *gin.Context) { controllers.Login(c, masterDB) })
	}

	auth := r.Group("/api")
	// auth.Use(middleware.VersionCheckMiddleware())
	auth.Use(middleware.JWTAuth(masterDB))
	auth.Use(middleware.RateLimiterMiddleware())
	auth.Use(middleware.TenantMiddleware(masterDB))
	{
		auth.POST("/logout", func(c *gin.Context) { controllers.Logout(c, masterDB) })
		auth.GET("/user", func(c *gin.Context) { controllers.GetUser(c, masterDB) })

		roleRoutes := auth.Group("/roles")
		{
			roleRoutes.POST("", middleware.RolePermissionMiddleware("role", "create", masterDB), func(c *gin.Context) { controllers.CreateRole(c, masterDB) })
			roleRoutes.GET("", middleware.RolePermissionMiddleware("role", "get", masterDB), func(c *gin.Context) { controllers.GetAllRoles(c, masterDB) })
			roleRoutes.GET("/list", middleware.RolePermissionMiddleware("role", "get", masterDB), func(c *gin.Context) { controllers.GetAllRolesList(c, masterDB) })
			roleRoutes.GET("/:id", middleware.RolePermissionMiddleware("role", "get", masterDB), func(c *gin.Context) { controllers.GetRoleByID(c, masterDB) })
			roleRoutes.GET("/:id/permissions", middleware.RolePermissionMiddleware("role", "get", masterDB), func(c *gin.Context) { controllers.GetPermissionsByRoleID(c, masterDB) })
			roleRoutes.PUT("/:id", middleware.RolePermissionMiddleware("role", "update", masterDB), func(c *gin.Context) { controllers.UpdateRole(c, masterDB) })
			roleRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("role", "delete", masterDB), func(c *gin.Context) { controllers.DeleteRole(c, masterDB) })
			roleRoutes.POST("/role/permissions/assign", middleware.RolePermissionMiddleware("role", "update", masterDB), func(c *gin.Context) { controllers.AssignPermissionsToRole(c, masterDB) })
			roleRoutes.POST("/role/permissions/remove", middleware.RolePermissionMiddleware("role", "update", masterDB), func(c *gin.Context) { controllers.RemovePermissionsFromRole(c, masterDB) })

		}

		userRoutes := auth.Group("/users")
		{
			userRoutes.POST("", middleware.RolePermissionMiddleware("user", "create", masterDB), func(c *gin.Context) { controllers.CreateUser(c, masterDB) })
			userRoutes.GET("", middleware.RolePermissionMiddleware("user", "get", masterDB), func(c *gin.Context) { controllers.GetAllUsers(c, masterDB) })
			userRoutes.GET("/:id", middleware.RolePermissionMiddleware("user", "get", masterDB), func(c *gin.Context) { controllers.GetUserByID(c, masterDB) })
			userRoutes.PUT("/:id", middleware.RolePermissionMiddleware("user", "update", masterDB), func(c *gin.Context) { controllers.UpdateUser(c, masterDB) })
			userRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("user", "delete", masterDB), func(c *gin.Context) { controllers.DeleteUser(c, masterDB) })
			userRoutes.POST("/change-password", middleware.StrictRateLimiterMiddleware(), func(c *gin.Context) { controllers.ChangePassword(c, masterDB) })
			userRoutes.GET("/my-permissions", func(c *gin.Context) { controllers.GetMyPermissions(c, masterDB) })
			userRoutes.POST("/check-permission", func(c *gin.Context) { controllers.CheckPermission(c, masterDB) })
		}

		moduleRoutes := auth.Group("/modules")
		{
			moduleRoutes.POST("", middleware.RolePermissionMiddleware("role", "create", masterDB), func(c *gin.Context) { controllers.CreateModule(c, masterDB) })
			moduleRoutes.GET("", middleware.RolePermissionMiddleware("role", "get", masterDB), func(c *gin.Context) { controllers.GetAllModules(c, masterDB) })
			moduleRoutes.GET("/:id", middleware.RolePermissionMiddleware("role", "get", masterDB), func(c *gin.Context) { controllers.GetModuleByID(c, masterDB) })
			moduleRoutes.PUT("/:id", middleware.RolePermissionMiddleware("role", "update", masterDB), func(c *gin.Context) { controllers.UpdateModule(c, masterDB) })
			moduleRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("role", "delete", masterDB), func(c *gin.Context) { controllers.DeleteModule(c, masterDB) })
			moduleRoutes.GET("/:id/permissions", middleware.RolePermissionMiddleware("role", "get", masterDB), func(c *gin.Context) { controllers.GetModulePermissions(c, masterDB) })
		}
		permissionRoutes := auth.Group("/permissions")
		{
			permissionRoutes.POST("", middleware.RolePermissionMiddleware("role", "create", masterDB), func(c *gin.Context) { controllers.CreatePermission(c, masterDB) })
			permissionRoutes.GET("", middleware.RolePermissionMiddleware("role", "get", masterDB), func(c *gin.Context) { controllers.GetAllPermissions(c, masterDB) })
			permissionRoutes.GET("/:id", middleware.RolePermissionMiddleware("role", "get", masterDB), func(c *gin.Context) { controllers.GetPermissionByID(c, masterDB) })
			permissionRoutes.PUT("/:id", middleware.RolePermissionMiddleware("role", "update", masterDB), func(c *gin.Context) { controllers.UpdatePermission(c, masterDB) })
			permissionRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("role", "delete", masterDB), func(c *gin.Context) { controllers.DeletePermission(c, masterDB) })
		}
	}
	return r

}
