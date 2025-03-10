package consts

const AppName = "go-auth"

const PasswordHashCost int = 10
const ErrExitStatus int = 2

// App consts
const (
	TelMinLength = 2 + 8  // 1+9
	TelMaxLength = 4 + 15 // 1+18

	EmailMinLength = 6

	PasswordMinLength = 8
	PasswordMaxLength = 50 // bcrypt 72

	PasscodeLength        = 8
	LongTextLength        = 32767 //  int(int16(^uint16(0) >> 1)) // equivalent of short.MaxValue
	DefaultTextLength     = 100
	DefaultMapZoom        = 12
	DefaultMaxQty         = 12
	TitleTextLengthTiny   = 12
	TitleTextLengthSmall  = 25
	TitleTextLengthInfo   = 35
	TitleTextLengthMedium = 50
	TitleTextLengthLarge  = 100

	// WF_STATUS_NEW       = 0
	// WF_STATUS_PROGRESS  = 6
	// WF_STATUS_DELETE    = 7
	// WF_STATUS_ERROR     = 10
	// WF_STATUS_SUCCESS   = 15
	// WF_STATUS_VOID      = 17
	// WF_STATUS_SIGNED    = 4
	// WF_STATUS_DELIVERED = 5
	// WF_STATUS_OUTBOX    = 3
	// WF_STATUS_READONLY  = 32
	// WF_STATUS_UNPAID    = 19
	// WF_STATUS_PAID      = 21
	// WF_STATUS_INQUEUE   = 31
)

// const (
// 	LogLevelError = 0
// 	LogLevelWarn  = 1
// 	LogLevelInfo  = 2
// 	LogLevelDebug = 3
// )

const (
	// PathAPI represents the group of PathAPI.
	PathAPI = "/api"
	// PathAuthSignin    = "/auth/signin"
	// PathAuthStatusAPI = "/auth/api/status"
	PathHome = "/"
)
const (
	RoleAdmin = "admin"
)

//nolint:gosec
const (
	PathSysMetricsAPI = "/sys/api/metrics"

	PathAuthHelloWorld = "/auth/hello-world"

	PathAuth             = "/auth"
	PathAuthAPI          = "/auth/api"
	PathAuthPingDebugAPI = "/auth/api/ping"
	//
	PathAuthAssets = "/auth/assets"

	PathAuthLockout        = "/auth/lockout"
	PathAuthAccessDenied   = "/auth/access-denied"
	PathAuthSignup         = "/auth/signup"
	PathAuthSignin         = "/auth/signin"
	PathAuthForgotPassword = "/auth/forgot-password"
	PathAuthSignout        = "/auth/signout"

	// PathAuthSignupTel = "/auth/signup/tel"
	// PathAuthSigninTel = "/auth/signin/tel"
	// PathAuthSignupEmail       = "/auth/signup/email"
	// PathAuthSigninEmail       = "/auth/signin/email"

	// PathAuthForgotPasswordTel    = "/auth/forgot-password/tel"
	// PathAuthForgotPasswordEmail          = "/auth/forgot-password/email"

	PathAuthForgotPasswordAPI      = "/auth/api/forgot-password"
	PathAuthSignupTelAPI           = "/auth/api/signup/tel"
	PathAuthSigninTelAPI           = "/auth/api/signin/tel"
	PathAuthForgotPasswordTelAPI   = "/auth/api/forgot-password/tel"
	PathAuthSignupEmailAPI         = "/auth/api/signup/email"
	PathAuthSigninEmailAPI         = "/auth/api/signin/email"
	PathAuthForgotPasswordEmailAPI = "/auth/api/forgot-password/email"
	PathAuthSignoutAPI             = "/auth/api/signout"
	PathAuthStatusAPI              = "/auth/api/status" // get _csrf, user related, no-cache
	PathAuthConfigAPI              = "/auth/api/config"

	PathAuthAccountSettings          = "/auth/account/settings"
	PathAuthAccountChangePasswordAPI = "/auth/api/account/change-password"
	PathAuthAccountDeleteDataAPI     = "/auth/api/account/delete-data"
)
