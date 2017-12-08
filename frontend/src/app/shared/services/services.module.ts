import { NgModule, Optional, SkipSelf } from "@angular/core";

import { ConfigService } from "./config/config.service";
import { PreloaderService } from "./preloader/preloader.service";
import { SpinnerService } from "./spinner/spinner.service";
import { DataService } from "./data/data.service";
import { ThemesService } from "./themes/themes.service";
import { ZonesService } from "app/shared/services/zones/zones.service";
import { SensorsService } from "./sensors/sensors.service";


@NgModule({
	imports: [],
	providers: [ConfigService, ThemesService, PreloaderService, SpinnerService, DataService, ZonesService, SensorsService],
	declarations: [],
	exports: []
})
export class ServicesModule {
	constructor(
		@Optional()
		@SkipSelf()
		parentModule: ServicesModule
	) { }
}
