import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { ZonesComponent } from "./zones.component";
import { SharedModule } from "app/shared/shared.module";
import { RouterModule } from "@angular/router";

const ROUTE = [
    { path: ":id", component: ZonesComponent },
];

@NgModule({
	  declarations: [
			ZonesComponent
		],
    imports: [
			CommonModule,
			SharedModule,
			RouterModule.forChild(ROUTE)
    ]
})
export class ZonesModule { }
