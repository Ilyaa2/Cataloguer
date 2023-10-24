package custom_errors

const (
	RecordNotFoundInCache      = "record not found in redistore"
	CantSetValueInCache        = "result not ok. Can't set value in redistore"
	IncorrectJsonStructure     = "incorrect structure of json in request"
	EmailNotRegistered         = "the user with this email doesn't exist"
	WrongPasswordOrEmail       = "presented password or email is wrong"
	IncorrectUsersFields       = "some fields are incorrect. \nDetails: "
	ThisEmailAlreadyRegistered = "user with this email already registered"
	UserDidntLogIn             = "you must be logged in"
	NotEnoughRights            = "you don't have rights to address to this resource"
	IncorrectMultipartFile     = "Error Retrieving the File, There's no 'item' or 'type' tag in formdata. \nDetails:"
	ServerSide                 = "error on the server"
	NoFilesFound               = "user don't have any files"
)
