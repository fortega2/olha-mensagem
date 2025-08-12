package dto

type UserDTO struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func NewUserDTO(id int64, username string) UserDTO {
	return UserDTO{
		ID:       id,
		Username: username,
	}
}
