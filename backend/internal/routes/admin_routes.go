package routes

import (
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/controllers"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/services"

	"timexeed/backend/internal/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
 * 〇 管理者用APIルート登録
 *
 * 管理者だけが使うAPIをここにまとめる。
 *
 * 重要：
 * 	管理者APIと従業員APIは完全に分離する。
 * 	従業員画面から使うAPIはここには書かない。
 *
 * ルール：
 * 	URLにIDを載せない。
 * 	targetUserId や attendanceId などは request body で受け取る。
 *
 * 管理者APIで扱うID：
 * 	操作している管理者本人のID
 * 		→ JWTから取得する
 *
 * 	操作対象のユーザーID
 * 		→ request body の targetUserId / targetUserIds で受け取る
 */
func RegisterAdminRoutes(r *gin.Engine, db *gorm.DB) {

	// 所属
	departmentBuilder := builders.NewDepartmentBuilder(db)
	departmentRepository := repositories.NewDepartmentRepository(db)
	departmentService := services.NewDepartmentService(departmentBuilder, departmentRepository)
	departmentController := controllers.NewDepartmentController(departmentService)

	// ユーザー
	userBuilder := builders.NewUserBuilder(db)
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userBuilder, userRepository)
	userController := controllers.NewUserController(userService)

	admin := r.Group("/admin")

	/*
	 * 管理者APIは、
	 * 1. JWT認証済みであること
	 * 2. role が ADMIN であること
	 * を必須にする。
	 */
	admin.Use(
		middlewares.AuthMiddleware(),
		middlewares.AdminMiddleware(),
	)

	{
		// 所属
		admin.POST("/departments/search", departmentController.SearchDepartments)
		admin.POST("/departments/detail", departmentController.GetDepartmentDetail)
		admin.POST("/departments/create", departmentController.CreateDepartment)
		admin.POST("/departments/update", departmentController.UpdateDepartment)
		admin.POST("/departments/delete", departmentController.DeleteDepartment)

		// ユーザー
		admin.POST("/users/search", userController.SearchUsers)
		admin.POST("/users/detail", userController.GetUserDetail)
		admin.POST("/users/create", userController.CreateUser)
		admin.POST("/users/update", userController.UpdateUser)
		admin.POST("/users/delete", userController.DeleteUser)
	}
}
