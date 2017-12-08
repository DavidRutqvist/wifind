import { Routes, RouterModule } from "@angular/router";
import { LayoutComponent } from "./layout.component";
const LAYOUT_ROUTES: Routes = [
	{
		path: "",
		component: LayoutComponent,
		children: [
			{ path: "", redirectTo: "dashboard", pathMatch: "full" }, {
				path: "dashboard",
				loadChildren: "../pages/dashboards/dashboards.module#DashboardsModule"
			}, {
				path: "management",
				loadChildren: "../pages/management/management.module#ManagementModule"
			}
		]
	},

	// 404 Page Not Found
	{ path: "**", redirectTo: "dashboard" }
];

export const LayoutRoutes = RouterModule.forChild(LAYOUT_ROUTES);
