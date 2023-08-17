package fork

type ProcessOpt = func(p *BackgroundProcess)

// WithEnv добавляет переменные окружения вида KEY=VALUE процессу
func WithEnv(env ...string) ProcessOpt {
	return func(p *BackgroundProcess) {
		p.cmd.Env = append(p.cmd.Env, env...)
	}
}

// WithArgs добавляет процессу аргументы командной строки
func WithArgs(args ...string) ProcessOpt {
	return func(p *BackgroundProcess) {
		p.cmd.Args = append(p.cmd.Args, args...)
	}
}
