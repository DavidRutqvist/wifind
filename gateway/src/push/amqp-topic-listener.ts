"use strict";
import * as amqp from "amqplib";
import * as Rx from "rxjs/Rx";
import * as log from "winston";

export abstract class AmqpTopicListener {
  private connection: amqp.Connection;
  private channel: amqp.Channel;
  private queue: string;

  constructor(private readonly connectionString: string, private readonly exchange: string, private readonly topicSubscriptions: string[]) { }

  protected abstract handle(message: amqp.Message): void;

  public connect(): Rx.Observable<void> {
    const connectionObservable = Rx.Observable.fromPromise(amqp.connect(this.connectionString))
      .share();
    
    connectionObservable.subscribe(connection => this.connection = connection);
    connectionObservable.subscribe(
      () => log.info("Connected to AMQP server"),
      err => log.error("Could not connect to AMQP server", err));

    const channelObservable = connectionObservable
      .flatMap(connection => connection.createChannel())
      .share();

    channelObservable.subscribe(channel => this.channel = channel);
    channelObservable.subscribe(
      () => log.info("Created channel"),
      err => log.error("Could not create channel", err));
    const exchangeObservable = channelObservable
      .flatMap(channel => channel.assertExchange(this.exchange, "topic"))
      .share();

    exchangeObservable
      .subscribe(
        () => log.info("Asserted exchange"),
        err => log.error("Could not assert exchange", err));

    const queueObservable = channelObservable
        .flatMap(channel => channel.assertQueue('', { exclusive: true }))
        .map(res => {
          this.queue = res.queue;
          return this.queue;
        })
        .share();

    queueObservable
      .subscribe(
        queue => log.info("Asserted queue: " + queue),
        err => log.error("Could not assert queue", err));

    const bindObservables: Rx.Observable<amqp.Replies.Empty>[] = [];
    for (let i = 0; i < this.topicSubscriptions.length; i++) {
      const bindObservable = queueObservable
        .combineLatest(channelObservable, (queue, channel) => channel)
        .flatMap(channel => channel.bindQueue(this.queue, this.exchange, this.topicSubscriptions[i]));

      bindObservables.push(bindObservable);
    }
    
    const bindWaitObservable = Rx.Observable.forkJoin(...bindObservables).share();

    bindWaitObservable.subscribe(
      res => log.info("Bound to %d topic subscriptions", res.length),
      err => log.error("Could not bind to topic subscriptions", err));

    // start listening
    bindWaitObservable
      .flatMap(() => this.listen())
      .subscribe(message => this.handle(message));

    return bindWaitObservable
      .map(() => this.noop());
  }

  private noop(): void {
    return;
  }

  private listen(): Rx.Observable<amqp.Message> {
    return Rx.Observable.create(observer => {
      // since arrow notation is used is this correct in this closure
      this.channel.consume(
        this.queue,
        msg => {
          observer.next(msg);
        },
        { noAck: true }
      );
    });
  }
  
  public close(): Rx.Observable<void> {
    log.info("Closing connection to AMQP server");

    const closeObservable = Rx.Observable.fromPromise(this.channel.close()).share();
    closeObservable.subscribe(() => log.info("AMQP connection closed"));

    return closeObservable;
  }
}
