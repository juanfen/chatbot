package schema

type expectedColumn struct {
	Name     string
	DataType string
	// Nullable: "YES" or "NO". Empty means skip checking nullable.
	Nullable string
}

type expectedTable struct {
	Columns []expectedColumn
}

var expectedSchemas = map[string]expectedTable{
	"sys_user": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "username", DataType: "varchar", Nullable: "NO"},
			{Name: "password_hash", DataType: "varchar", Nullable: "NO"},
			{Name: "nickname", DataType: "varchar", Nullable: "NO"},
			{Name: "status", DataType: "integer", Nullable: "NO"},
			{Name: "is_admin", DataType: "integer", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "deleted", DataType: "integer", Nullable: "NO"},
		},
	},
	"sys_role": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "code", DataType: "varchar", Nullable: "NO"},
			{Name: "name", DataType: "varchar", Nullable: "NO"},
			{Name: "status", DataType: "integer", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "deleted", DataType: "integer", Nullable: "NO"},
		},
	},
	"sys_user_role": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "user_id", DataType: "varchar", Nullable: "NO"},
			{Name: "role_id", DataType: "varchar", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "deleted", DataType: "integer", Nullable: "NO"},
		},
	},
	"sys_route": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "parent_id", DataType: "varchar", Nullable: "NO"},
			{Name: "name", DataType: "varchar", Nullable: "NO"},
			{Name: "path", DataType: "varchar", Nullable: "NO"},
			{Name: "component", DataType: "varchar", Nullable: "NO"},
			{Name: "title", DataType: "varchar", Nullable: "NO"},
			{Name: "i18n_key", DataType: "varchar", Nullable: "NO"},
			{Name: "icon", DataType: "varchar", Nullable: "NO"},
			{Name: "order_no", DataType: "integer", Nullable: "NO"},
			{Name: "hide_in_menu", DataType: "integer", Nullable: "NO"},
			{Name: "keep_alive", DataType: "integer", Nullable: "NO"},
			{Name: "status", DataType: "integer", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "deleted", DataType: "integer", Nullable: "NO"},
		},
	},
	"sys_role_route": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "role_id", DataType: "varchar", Nullable: "NO"},
			{Name: "route_id", DataType: "varchar", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "deleted", DataType: "integer", Nullable: "NO"},
		},
	},
	"biz_announcement": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "title", DataType: "varchar", Nullable: "NO"},
			{Name: "content", DataType: "varchar", Nullable: "NO"},
			{Name: "sort_no", DataType: "integer", Nullable: "NO"},
			{Name: "status", DataType: "integer", Nullable: "NO"},
			{Name: "publish_time", DataType: "timestamp", Nullable: "YES"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "deleted", DataType: "integer", Nullable: "NO"},
		},
	},
	"biz_package": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "code", DataType: "varchar", Nullable: "NO"},
			{Name: "name", DataType: "varchar", Nullable: "NO"},
			{Name: "description", DataType: "varchar", Nullable: "NO"},
			{Name: "quota_calls", DataType: "bigint", Nullable: "NO"},
			{Name: "sort_no", DataType: "integer", Nullable: "NO"},
			{Name: "status", DataType: "integer", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "deleted", DataType: "integer", Nullable: "NO"},
		},
	},
	"biz_usage_log": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "user_id", DataType: "varchar", Nullable: "NO"},
			{Name: "channel", DataType: "varchar", Nullable: "NO"},
			{Name: "ref_text", DataType: "varchar", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
		},
	},
	"biz_user_quota": {
		Columns: []expectedColumn{
			{Name: "user_id", DataType: "varchar", Nullable: "NO"},
			{Name: "package_id", DataType: "varchar", Nullable: "NO"},
			{Name: "package_title", DataType: "varchar", Nullable: "NO"},
			{Name: "total_quota", DataType: "bigint", Nullable: "NO"},
			{Name: "used_quota", DataType: "bigint", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
		},
	},
	"biz_user_preference": {
		Columns: []expectedColumn{
			{Name: "user_id", DataType: "varchar", Nullable: "NO"},
			{Name: "standard_text", DataType: "varchar", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
		},
	},
	"biz_user_voice": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "user_id", DataType: "varchar", Nullable: "NO"},
			{Name: "file_name", DataType: "varchar", Nullable: "NO"},
			{Name: "upload_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "duration_seconds", DataType: "double precision", Nullable: "NO"},
			{Name: "is_active", DataType: "integer", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "deleted", DataType: "integer", Nullable: "NO"},
		},
	},
	"biz_device_audio": {
		Columns: []expectedColumn{
			{Name: "id", DataType: "varchar", Nullable: "NO"},
			{Name: "mac_id", DataType: "varchar", Nullable: "NO"},
			{Name: "audio_blob", DataType: "bytea", Nullable: "NO"},
			{Name: "audio_len", DataType: "varchar", Nullable: "NO"},
			{Name: "context", DataType: "text", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "expire_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "delete_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "speaker_id", DataType: "varchar", Nullable: "YES"},
			{Name: "bound_user_id", DataType: "varchar", Nullable: "YES"},
		},
	},
	"biz_user_speaker": {
		Columns: []expectedColumn{
			{Name: "speaker_id", DataType: "varchar", Nullable: "NO"},
			{Name: "user_id", DataType: "varchar", Nullable: "NO"},
			{Name: "speaker_name", DataType: "varchar", Nullable: "NO"},
			{Name: "personal_lines", DataType: "text", Nullable: "NO"},
			{Name: "create_time", DataType: "timestamp", Nullable: "NO"},
			{Name: "update_time", DataType: "timestamp", Nullable: "NO"},
		},
	},
}

