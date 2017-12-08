import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { InstallUnitComponent } from "./install-unit/install-unit.component";
import { RouterModule } from "@angular/router";
import { SharedModule } from "../../shared/shared.module";

const ROUTES = [
    { path: "install-unit", component: InstallUnitComponent },
];

@NgModule({
	  declarations: [
			InstallUnitComponent
		],
    imports: [
			CommonModule,
			SharedModule,
			RouterModule.forChild(ROUTES)
    ]
})
export class ManagementModule { }
