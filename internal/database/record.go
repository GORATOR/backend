package database

const (
	activeRecordWhere = "id = ? and active = true"
)

func GetRecord(id uint, entity interface{}) error {
	result := postgresConnection.Where(activeRecordWhere, id).First(&entity)
	return result.Error
}

func DisableRecord(id uint, entity interface{}) error {
	result := postgresConnection.Model(&entity).Where(activeRecordWhere, id).Update("active", false)
	return result.Error
}

func EnableRecord(id uint, entity interface{}) error {
	result := postgresConnection.Model(&entity).Where("id = ? and active = false", id).Update("active", true)
	return result.Error
}
