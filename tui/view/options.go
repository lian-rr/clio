package view

type OptFunc func(main *Main)

func WithTeacher(teacher teacher) OptFunc {
	return func(main *Main) {
		main.teacher = teacher
	}
}
