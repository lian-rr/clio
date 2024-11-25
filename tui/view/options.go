package view

type OptFunc func(main *Main)

func WithProfessor(professor professor) OptFunc {
	return func(main *Main) {
		main.professor = professor
	}
}
