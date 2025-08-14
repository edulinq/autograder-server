package core

import (
	"reflect"
	"regexp"

	"github.com/edulinq/autograder/internal/model"
)

var minRoleRegex = regexp.MustCompile(`^Min(Server|Course)Role\w+$`)

// The minimum server roles required encoded as a type so it can be embedded into a request struct.
// Using any of these implies your request is at least an APIRequestUserContext.

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-server-role-root" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum server role of root to complete this operation."
//	}
type MinServerRoleRoot bool

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-server-role-owner" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum server role of owner to complete this operation."
//	}
type MinServerRoleOwner bool

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-server-role-admin" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum server role of admin to complete this operation."
//	}
type MinServerRoleAdmin bool

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-server-role-course-creator" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum server role of course creator to complete this operation."
//	}
type MinServerRoleCourseCreator bool

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-server-role-user" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum server role of user to complete this operation."
//	}
type MinServerRoleUser bool

// The minimum course roles required encoded as a type so it can be embedded into a request struct.
// Using any of these implies your request is at least an APIRequestCourseUserContext.

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-course-role-owner" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum course role of owner to complete this operation."
//	}
type MinCourseRoleOwner bool

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-course-role-admin" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum course role of admin to complete this operation."
//	}
type MinCourseRoleAdmin bool

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-course-role-grader" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum course role of grader to complete this operation."
//	}
type MinCourseRoleGrader bool

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-course-role-student" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum course role of student to complete this operation."
//	}
type MinCourseRoleStudent bool

//	__TYPE_DESCRIPTION_OVERRIDE__: "min-course-role-other" = {
//	    "category": "role",
//	    "description": "The requesting user must have a minimum course role of other to complete this operation."
//	}
type MinCourseRoleOther bool

// Take a request (or any object),
// go through all the fields and look for fields typed as the encoded MinServerRole* fields.
// Return the maximum amongst the found roles.
// Return: (course role, found role).
func getMaxServerRole(request any) (model.ServerUserRole, bool) {
	reflectValue := reflect.ValueOf(request)

	// Dereference any pointer.
	if reflectValue.Kind() == reflect.Pointer {
		reflectValue = reflectValue.Elem()
	}

	foundRole := false
	role := model.ServerRoleUnknown

	for i := 0; i < reflectValue.NumField(); i++ {
		fieldValue := reflectValue.Field(i)

		if fieldValue.Type() == reflect.TypeOf((*MinServerRoleRoot)(nil)).Elem() {
			foundRole = true
			if role < model.ServerRoleRoot {
				role = model.ServerRoleRoot
			}
		} else if fieldValue.Type() == reflect.TypeOf((*MinServerRoleOwner)(nil)).Elem() {
			foundRole = true
			if role < model.ServerRoleOwner {
				role = model.ServerRoleOwner
			}
		} else if fieldValue.Type() == reflect.TypeOf((*MinServerRoleAdmin)(nil)).Elem() {
			foundRole = true
			if role < model.ServerRoleAdmin {
				role = model.ServerRoleAdmin
			}
		} else if fieldValue.Type() == reflect.TypeOf((*MinServerRoleCourseCreator)(nil)).Elem() {
			foundRole = true
			if role < model.ServerRoleCourseCreator {
				role = model.ServerRoleCourseCreator
			}
		} else if fieldValue.Type() == reflect.TypeOf((*MinServerRoleUser)(nil)).Elem() {
			foundRole = true
			if role < model.ServerRoleUser {
				role = model.ServerRoleUser
			}
		}
	}

	return role, foundRole
}

// Take a request (or any object),
// go through all the fields and look for fields typed as the encoded MinCourseRole* fields.
// Return the maximum amongst the found roles.
// Return: (course role, found role).
func getMaxCourseRole(request any) (model.CourseUserRole, bool) {
	reflectValue := reflect.ValueOf(request)

	// Dereference any pointer.
	if reflectValue.Kind() == reflect.Pointer {
		reflectValue = reflectValue.Elem()
	}

	foundRole := false
	role := model.CourseRoleUnknown

	for i := 0; i < reflectValue.NumField(); i++ {
		fieldValue := reflectValue.Field(i)

		if fieldValue.Type() == reflect.TypeOf((*MinCourseRoleOwner)(nil)).Elem() {
			foundRole = true
			if role < model.CourseRoleOwner {
				role = model.CourseRoleOwner
			}
		} else if fieldValue.Type() == reflect.TypeOf((*MinCourseRoleAdmin)(nil)).Elem() {
			foundRole = true
			if role < model.CourseRoleAdmin {
				role = model.CourseRoleAdmin
			}
		} else if fieldValue.Type() == reflect.TypeOf((*MinCourseRoleGrader)(nil)).Elem() {
			foundRole = true
			if role < model.CourseRoleGrader {
				role = model.CourseRoleGrader
			}
		} else if fieldValue.Type() == reflect.TypeOf((*MinCourseRoleStudent)(nil)).Elem() {
			foundRole = true
			if role < model.CourseRoleStudent {
				role = model.CourseRoleStudent
			}
		} else if fieldValue.Type() == reflect.TypeOf((*MinCourseRoleOther)(nil)).Elem() {
			foundRole = true
			if role < model.CourseRoleOther {
				role = model.CourseRoleOther
			}
		}
	}

	return role, foundRole
}
