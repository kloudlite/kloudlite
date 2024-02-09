package handler

import (
	"fmt"

	"github.com/getlantern/systray"
	"github.com/google/uuid"
	ns "github.com/kloudlite/kl/app/handler/name-conts"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func (h *handler) RedrawEnvs(m projectMap) {
	// map h.projects from m
	currentVersion := uuid.New().String()
	h.projectRenderVersion = currentVersion

	eSwitchBtn := h.itemMap[ns.EnvBtn]

	fmt.Printf("%+v\n", m)

	pIndex := 0
	for k, v := range m {
		if pIndex < len(h.projects) {
			fmt.Println("updating existing project")
			prj := h.projects[pIndex]

			prj.projectBtn.SetTitle(k)
			prj.projectBtn.SetTooltip(fmt.Sprintf("switch to %s", k))
			prj.projectBtn.Show()

			var envs []env
			eIndex := 0

			for _, envName := range v {
				if eIndex < len(prj.envs) {
					fmt.Println("updating existing env")
					env := prj.envs[eIndex]
					env.envBtn.SetTitle(envName)
					env.envBtn.SetTooltip(fmt.Sprintf("switch to %s", envName))
					env.envBtn.Show()
				} else {
					fmt.Println("adding new env")
					envs = append(envs, env{
						name:   envName,
						envBtn: prj.projectBtn.AddSubMenuItem(envName, fmt.Sprintf("switch to %s", envName)),
					})
				}

				eIndex++
			}

			fmt.Println(len(prj.envs), len(v), "env")

			if len(prj.envs) > len(v) {
				fmt.Println("hiding env p")
				for i := len(v); i < len(prj.envs); i++ {
					fmt.Println("hiding env")
					prj.envs[i].envBtn.Hide()
				}
			}

			prj.envs = envs

			h.projects[pIndex] = prj

			pIndex++
		} else {
			fmt.Println("adding new project")
			h.projects = append(h.projects, project{
				name:       k,
				projectBtn: eSwitchBtn.AddSubMenuItem(k, fmt.Sprintf("switch to %s", v)),
				envs: func() []env {
					var envs []env

					for _, envName := range v {
						envs = append(envs, env{
							name:   envName,
							envBtn: eSwitchBtn.AddSubMenuItem(envName, fmt.Sprintf("switch to %s", envName)),
						})
					}

					return envs
				}(),
			})
		}
	}

	if len(h.projects) > len(m) {
		fmt.Println("hiding project")
		for i := len(m); i < len(h.projects); i++ {
			h.projects[i].projectBtn.Hide()
		}
	}
}

func (h *handler) ReconEnv() {
	var envTitle, envBtn *systray.MenuItem
	if h.itemMap[ns.EnvTitle] != nil {
		envTitle = h.itemMap[ns.EnvTitle]
	} else {
		envTitle = systray.AddMenuItem("Environment", "")
		h.AddItem(ns.EnvTitle, envTitle)
		envTitle.Disable()
	}

	if h.itemMap[ns.EnvBtn] != nil {
		envBtn = h.itemMap[ns.EnvBtn]
		envBtn.SetTitle("...")
	} else {
		envBtn = systray.AddMenuItem("...", "")
		h.AddItem(ns.EnvBtn, envBtn)
	}

	go func() {
		projects := projectMap{}
		ps, err := server.ListProjects()
		if err != nil {
			fn.PrintError(err)
			fn.Notify("error:", err)
		}

		for _, p := range ps {
			projects[p.Metadata.Name] = map[string]string{}
			e, err := server.ListEnvs(fn.MakeOption("projectName", p.Metadata.Name))
			if err != nil {
				fn.PrintError(err)
				fn.Notify("error:", err)
			}

			for _, env := range e {
				projects[p.Metadata.Name][env.Metadata.Name] = env.Metadata.Name
			}
		}

		h.RedrawEnvs(projects)
	}()

	go func() {
		d, err := server.GetDevice()
		if err != nil {
			fn.PrintError(err)
			fn.Notify("error:", err)

			envBtn.SetTitle("...")
			return
		}

		envBtn.SetTitle(fmt.Sprintf("%s/%s", d.ProjectName, d.EnvName))
	}()

}
