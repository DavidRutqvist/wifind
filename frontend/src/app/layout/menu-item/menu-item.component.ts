import { Component, OnInit, Input } from '@angular/core';

@Component({
  selector: 'app-menu-item',
  templateUrl: './menu-item.component.html',
  styleUrls: ['./menu-item.component.scss']
})
export class MenuItemComponent implements OnInit {
  @Input() item: MenuItem;

  constructor() { }

  ngOnInit() {
  }

}

export interface MenuItem {
  label: string;
  link?: string;
  icon?: string;
  children?: MenuItem[];
}
