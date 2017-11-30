import { Routes, RouterModule } from "@angular/router";
import { LayoutComponent } from "./layout.component";
const LAYOUT_ROUTES: Routes = [
	{
		path: "",
		component: LayoutComponent,
		children: [
			{ path: "", redirectTo: "dashboard", pathMatch: "full" },
			//---------------------------------------------------------->
			//Dashboard
			//---------------------------------------------------------->
			{
				path: "dashboard",
				loadChildren: "../pages/dashboards/dashboards.module#DashboardsModule"
			}
		]
	},

	// 404 Page Not Found
	{ path: "**", redirectTo: "dashboard" }
];

export const LayoutRoutes = RouterModule.forChild(LAYOUT_ROUTES);
