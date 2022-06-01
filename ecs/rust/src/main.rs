use bevy_ecs::prelude::*;
use rand::prelude::*;
use std::time::{Instant};


#[derive(Component)]
#[derive(Debug)]
struct Position { x: f64, y: f64 }

#[derive(Component)]
#[derive(Debug)]
struct Velocity { x: f64, y: f64 }

#[derive(Component)]
#[derive(Debug)]
struct Collider {
    radius: f64,
}

#[derive(Component)]
#[derive(Debug)]
struct Count {
    count: i32,
}

const size : i64 = 1000;
const iterations : i64 = 1000;
const maxSpeed : f64 = 10.0;
const maxCollider : f64 = 1.0;
const maxPosition : f64 = 100.0;
const collisionLimit :i32 = 100;

fn main() {
    println!("starting");

    let mut world = World::default();

    // For loop
    let mut rng = rand::thread_rng();
    for _i in 0..size {
        world.spawn()
            .insert(Position{ x: maxPosition * rng.gen::<f64>(), y: maxPosition * rng.gen::<f64>() })
            .insert(Velocity{ x: maxSpeed * rng.gen::<f64>(), y: maxSpeed * rng.gen::<f64>() })
            .insert(Collider{ radius: maxCollider * rng.gen::<f64>() })
            .insert(Count{ count: 0 });
    }

    let mut schedule = Schedule::default();

    // Stages
    schedule.add_stage("update", SystemStage::single_threaded()
                       .with_system(update_position)
    );
    schedule.add_stage("collision", SystemStage::single_threaded()
                       .with_system(check_collision)
    );
/*    schedule.add_stage("print", SystemStage::single_threaded()
                       .with_system(print_count)
    );*/

    for i in 0..iterations {
        let start = Instant::now();
        schedule.run(&mut world);
        let duration = start.elapsed();
        println!("{}, {:?}", i, duration)
    }
}

// https://bevy-cheatbook.github.io/programming/paramset.html
fn check_collision(mut commands: Commands, mut query: Query<(Entity, &Position, &Collider, &mut Count)>, query2: Query<(Entity, &Position, &Collider)>) {
    let mut deathCount = 0;
    for (ent1, position, collider, mut count) in query.iter_mut() {
        for (ent2, targPos, targCollider) in query2.iter() {
            if ent1 == ent2 {
                continue;
            }

            let dx = position.x - targPos.x;
            let dy = position.y - targPos.y;
            let distSq = (dx * dx) + (dy * dy);

            let dr = collider.radius * targCollider.radius;
            let drSq = dr * dr;

            if drSq > distSq {
                count.count += 1;
            }

            // TODO move to outer loop?
            if collisionLimit > 0 && count.count > collisionLimit {
                deathCount += 1;
                commands.entity(ent1).despawn();
                break;
            }
        }
    }

    let mut rng = rand::thread_rng();
    for _i in 0..deathCount {
        commands.spawn()
            .insert(Position{ x: maxPosition * rng.gen::<f64>(), y: maxPosition * rng.gen::<f64>() })
            .insert(Velocity{ x: maxSpeed * rng.gen::<f64>(), y: maxSpeed * rng.gen::<f64>() })
            .insert(Collider{ radius: maxCollider * rng.gen::<f64>() })
            .insert(Count{ count: 0 });
    }
}

fn update_position(mut query: Query<(&mut Position, &mut Velocity)>) {
    let fixed_time = 0.015;

    for (mut position, mut velocity) in query.iter_mut() {
        position.x += velocity.x * fixed_time;
        position.y += velocity.y * fixed_time;

        // Bump into the bounding rect
        if position.x <= 0.0 || position.x >= maxPosition {
            velocity.x = -velocity.x;
        }
        if position.y <= 0.0 || position.y >= maxPosition {
            velocity.y = -velocity.y;
        }
    }
}

/*fn print_count(query: Query<&Count>) {
    for count in query.iter() {
        println!("count: {:?}", count);
    }
}*/
