import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ZoneBadgeComponent } from './zone-badge.component';

describe('ZoneBadgeComponent', () => {
  let component: ZoneBadgeComponent;
  let fixture: ComponentFixture<ZoneBadgeComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ZoneBadgeComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ZoneBadgeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });
});
