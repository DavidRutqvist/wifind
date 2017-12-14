import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { RealtimeGraphComponent } from './realtime-graph.component';

describe('RealtimeGraphComponent', () => {
  let component: RealtimeGraphComponent;
  let fixture: ComponentFixture<RealtimeGraphComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ RealtimeGraphComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RealtimeGraphComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });
});
