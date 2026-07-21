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
		public.POST("/customer/register", middleware.StrictRateLimiterMiddleware(), func(c *gin.Context) { controllers.RegisterCustomer(c, masterDB) })
		public.POST("/customer/login", middleware.StrictRateLimiterMiddleware(), func(c *gin.Context) { controllers.LoginCustomer(c, masterDB) })
	}

	customerAuth := r.Group("/api/customer")
	customerAuth.Use(middleware.CustomerJWTAuth(masterDB))
	customerAuth.Use(middleware.RateLimiterMiddleware())
	{
		customerAuth.GET("/me", func(c *gin.Context) { controllers.GetCustomerProfile(c, masterDB) })
		customerAuth.PUT("/me", func(c *gin.Context) { controllers.UpdateCustomerProfile(c, masterDB) })
		customerAuth.GET("/addresses", func(c *gin.Context) { controllers.GetMyCustomerAddresses(c, masterDB) })
		customerAuth.POST("/addresses", func(c *gin.Context) { controllers.CreateMyCustomerAddress(c, masterDB) })
		customerAuth.PUT("/addresses/:address_id", func(c *gin.Context) { controllers.UpdateMyCustomerAddress(c, masterDB) })
		customerAuth.DELETE("/addresses/:address_id", func(c *gin.Context) { controllers.DeleteMyCustomerAddress(c, masterDB) })
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

		customerRoutes := auth.Group("/customers")
		{
			customerRoutes.POST("", middleware.RolePermissionMiddleware("customer", "create", masterDB), func(c *gin.Context) { controllers.CreateCustomer(c, masterDB) })
			customerRoutes.GET("", middleware.RolePermissionMiddleware("customer", "get", masterDB), func(c *gin.Context) { controllers.GetAllCustomers(c, masterDB) })
			customerRoutes.GET("/:id", middleware.RolePermissionMiddleware("customer", "get", masterDB), func(c *gin.Context) { controllers.GetCustomerByID(c, masterDB) })
			customerRoutes.PUT("/:id", middleware.RolePermissionMiddleware("customer", "update", masterDB), func(c *gin.Context) { controllers.UpdateCustomer(c, masterDB) })
			customerRoutes.PATCH("/:id/status", middleware.RolePermissionMiddleware("customer", "update", masterDB), func(c *gin.Context) { controllers.UpdateCustomerStatus(c, masterDB) })
			customerRoutes.POST("/:id/set-password", middleware.RolePermissionMiddleware("customer", "update", masterDB), func(c *gin.Context) { controllers.SetCustomerPassword(c, masterDB) })
			customerRoutes.DELETE("/:id", middleware.RolePermissionMiddleware("customer", "delete", masterDB), func(c *gin.Context) { controllers.DeleteCustomer(c, masterDB) })
			customerRoutes.POST("/:id/addresses", middleware.RolePermissionMiddleware("customer_address", "create", masterDB), func(c *gin.Context) { controllers.CreateCustomerAddress(c, masterDB) })
			customerRoutes.GET("/:id/addresses", middleware.RolePermissionMiddleware("customer_address", "get", masterDB), func(c *gin.Context) { controllers.GetCustomerAddresses(c, masterDB) })
			customerRoutes.PUT("/:id/addresses/:address_id", middleware.RolePermissionMiddleware("customer_address", "update", masterDB), func(c *gin.Context) { controllers.UpdateCustomerAddress(c, masterDB) })
			customerRoutes.DELETE("/:id/addresses/:address_id", middleware.RolePermissionMiddleware("customer_address", "delete", masterDB), func(c *gin.Context) { controllers.DeleteCustomerAddress(c, masterDB) })
		}

		nutriSunRoutes := auth.Group("/nutrisun")
		{
			nutriSunRoutes.POST("/bootstrap-modules", middleware.RolePermissionMiddleware("role", "update", masterDB), func(c *gin.Context) { controllers.BootstrapNutriSunModules(c, masterDB) })
		}
	}
	return r

}
