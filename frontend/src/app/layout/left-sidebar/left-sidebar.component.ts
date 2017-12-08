import {
	Component,
	ViewEncapsulation,
	OnInit,
	trigger,
	state,
	style,
	transition,
	animate,
	ElementRef,
	HostListener
} from "@angular/core";
import { GlobalState } from "../../app.state";
import { ConfigService } from "../../shared/services/config/config.service";
import { ZonesService } from "app/shared/services/zones/zones.service";
import { Zone } from "app/shared/services/zones/zone";
import { MenuItem } from "app/layout/menu-item/menu-item.component";

@Component({
	selector: "app-sidebar",
	templateUrl: "./left-sidebar.component.html",
	styleUrls: ["./left-sidebar.component.scss"]
})
export class LeftSidebarComponent implements OnInit {
	public scrollbarOptions = {
		axis: "y",
		theme: "minimal",
		scrollInertia: 0,
		mouseWheel: { preventDefault: true }
	};

	zonesIsLoading = false;
	menuItems: MenuItem[] = [
		{
			label: "Dashboard",
			link: "/dashboard",
			icon: "zmdi zmdi-view-dashboard"
		},
		{
			label: "Management",
			icon: "zmdi zmdi-settings",
			children: [
				{
					label: "Install Unit",
					link: "/management/install-unit"
				}
			]
		}
	];

	constructor(
		public config: ConfigService,
		private _elementRef: ElementRef,
		private _state: GlobalState,
		private zonesSvc: ZonesService) { }

	ngOnInit() {
		this.zonesIsLoading = true;
		this.zonesSvc.getTopZones()
			.finally(() => this.zonesIsLoading = false)
			.subscribe(zones => this.appendZonesToMenu(zones));
	}

	private appendZonesToMenu(zones: Zone[]): void {
		if (zones.length > 0) {
			this.menuItems.push({
				label: "Zones"
			});
		}

		for (const zone of zones) {
			const menuItem: MenuItem = this.getMenuItemFromZone(zone, "zmdi zmdi-pin");

			this.menuItems.push(menuItem);
			this.appendZoneChildren(menuItem, zone);
		}
	}

	private appendZoneChildren(menuItem: MenuItem, zone: Zone): void {
		this.zonesSvc.getZoneChildren(zone.id)
			.subscribe(children => this.appendZoneChildrenToMenu(menuItem, children));
	}

	private appendZoneChildrenToMenu(menuItem: MenuItem, children: Zone[]): void {
		if (children.length > 0) {
			menuItem.children = [];

			for (const child of children) {
				const subMenuItem: MenuItem = this.getMenuItemFromZone(child);
				menuItem.children.push(subMenuItem);
				this.appendZoneChildren(subMenuItem, child);
			}
		}
	}

	private getMenuItemFromZone(zone: Zone, icon?: string): MenuItem {
		return {
			label: zone.name,
			link: "/zones/" + zone.id,
			icon: icon
		};
	}

	@HostListener("window:resize")
	public onWindowResize(): void { }
}
