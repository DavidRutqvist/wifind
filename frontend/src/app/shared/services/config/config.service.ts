import { Injectable } from "@angular/core";
import axios, { AxiosInstance } from "axios";
declare var $: any;

@Injectable()
export class ConfigService {
  public app: any;
  public appLayout: any;
  public breakpoint: any;

  constructor() {
    this.app = {
			api: "http://api.wifind.se:9999/api"
		};

		this.appLayout = {
			isApp_Boxed: false,
			isApp_SidebarLeftCollapsed: false,
			isApp_MobileSidebarLeftOpen: false,
			isApp_SidebarRightOpen: false,
			isApp_BackdropVisible: false
		};

		this.breakpoint = {
			desktopLG: 1280,
			desktop: 992,
			tablet: 768,
			mobile: 576
		};
	}

	public getApiClient(): AxiosInstance {
		return axios.create({
			baseURL: this.app.api
		});
	}
}
