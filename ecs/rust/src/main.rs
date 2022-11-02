use bevy_ecs::prelude::*;
use rand::prelude::*;
use std::env;
use std::time::Instant;

mod alt;

#[derive(Component, Debug)]
struct Position {
    x: f64,
    y: f64,
}

#[derive(Component, Debug)]
struct Velocity {
    x: f64,
    y: f64,
}

#[derive(Component, Debug)]
struct Collider {
    radius: f64,
}

#[derive(Component, Debug)]
struct Count {
    count: i32,
}

const ITERATIONS: i64 = 50;

const MAX_POSITION: f64 = 100.0;
const MAX_SPEED: f64 = 10.0;
const MAX_COLLIDER: f64 = 1.0;

fn main() {
    let args: Vec<String> = env::args().collect();
    let program = args[1].to_string();
    let size = args[2].parse::<i32>().unwrap();
    let collision_limit = args[3].parse::<i32>().unwrap();

    println!("Iter Time");
    if program == "ecs" {
        bevy(size, collision_limit);
    } else if program == "native" {
        native(size, collision_limit);
    } else if program == "nativeSplit" {
        native_split(size, collision_limit);
    } else if program == "fennewald" {
        alt::native(size as usize, collision_limit as usize);
    }
}

fn bevy(size: i32, collision_limit: i32) {
    //    println!("starting");

    let mut world = World::default();

    // For loop
    let mut rng = rand::thread_rng();
    for _i in 0..size {
        world
            .spawn()
            .insert(Position {
                x: MAX_POSITION * rng.gen::<f64>(),
                y: MAX_POSITION * rng.gen::<f64>(),
            })
            .insert(Velocity {
                x: MAX_SPEED * rng.gen::<f64>(),
                y: MAX_SPEED * rng.gen::<f64>(),
            })
            .insert(Collider {
                radius: MAX_COLLIDER * rng.gen::<f64>(),
            })
            .insert(Count { count: 0 });
    }

    let mut schedule = Schedule::default();

    // Stages
    schedule.add_stage(
        "update",
        SystemStage::single_threaded().with_system(update_position),
    );

    let collision = move |commands: Commands,
                          query: Query<(Entity, &Position, &Collider, &mut Count)>,
                          query2: Query<(Entity, &Position, &Collider)>| {
        check_collision(commands, query, query2, collision_limit);
    };
    schedule.add_stage(
        "collision",
        SystemStage::single_threaded().with_system(collision),
    );
    /*    schedule.add_stage("print", SystemStage::single_threaded()
                       .with_system(print_count)
    );*/

    for i in 0..ITERATIONS {
        let start = Instant::now();
        schedule.run(&mut world);
        let duration = start.elapsed();
        println!("{} {:?}", i, (duration.as_micros() as f64) / 1000000.0)
    }
}

// https://bevy-cheatbook.github.io/programming/paramset.html
fn check_collision(
    mut commands: Commands,
    mut query: Query<(Entity, &Position, &Collider, &mut Count)>,
    query2: Query<(Entity, &Position, &Collider)>,
    collision_limit: i32,
) {
    let mut death_count = 0;

    for (ent1, position, collider, mut count) in query.iter_mut() {
        for (ent2, targ_pos, targ_col) in query2.iter() {
            if ent1 == ent2 {
                continue;
            }

            let dx = position.x - targ_pos.x;
            let dy = position.y - targ_pos.y;
            let dist_squared = (dx * dx) + (dy * dy);

            let dr = collider.radius + targ_col.radius;
            let dr_squared = dr * dr;

            if dr_squared > dist_squared {
                count.count += 1;
            }

            // TODO move to outer loop?
            if collision_limit > 0 && count.count > collision_limit {
                death_count += 1;
                commands.entity(ent1).despawn();
                break;
            }
        }
    }

    //    println!("DeathCount: {}", death_count);
    let mut rng = rand::thread_rng();
    for _i in 0..death_count {
        commands
            .spawn()
            .insert(Position {
                x: MAX_POSITION * rng.gen::<f64>(),
                y: MAX_POSITION * rng.gen::<f64>(),
            })
            .insert(Velocity {
                x: MAX_SPEED * rng.gen::<f64>(),
                y: MAX_SPEED * rng.gen::<f64>(),
            })
            .insert(Collider {
                radius: MAX_COLLIDER * rng.gen::<f64>(),
            })
            .insert(Count { count: 0 });
    }

    //    let mut cnt = 0;
    //    for (_ent1, _position, _collider, _count) in query.iter() {
    //        cnt += 1;
    //    }
    //    println!("Ent Count: {}", cnt);
}

fn update_position(mut query: Query<(&mut Position, &mut Velocity)>) {
    let fixed_time = 0.015;

    for (mut position, mut velocity) in query.iter_mut() {
        position.x += velocity.x * fixed_time;
        position.y += velocity.y * fixed_time;

        // Bump into the bounding rect
        if position.x <= 0.0 || position.x >= MAX_POSITION {
            velocity.x = -velocity.x;
        }
        if position.y <= 0.0 || position.y >= MAX_POSITION {
            velocity.y = -velocity.y;
        }
    }
}

/*fn print_count(query: Query<&Count>) {
    for count in query.iter() {
        println!("count: {:?}", count);
    }
}*/

fn native(size: i32, collision_limit: i32) {
    let mut ids = Vec::new();
    let mut pos = Vec::new();
    let mut vel = Vec::new();
    let mut col = Vec::new();
    let mut cnt = Vec::new();

    let mut rng = rand::thread_rng();
    for i in 0..size {
        ids.push(i + 2);
        pos.push(Position {
            x: MAX_POSITION * rng.gen::<f64>(),
            y: MAX_POSITION * rng.gen::<f64>(),
        });
        vel.push(Velocity {
            x: MAX_SPEED * rng.gen::<f64>(),
            y: MAX_SPEED * rng.gen::<f64>(),
        });
        col.push(Collider {
            radius: MAX_COLLIDER * rng.gen::<f64>(),
        });
        cnt.push(Count { count: 0 });
    }

    let fixed_time = 0.015;

    for iter_count in 0..ITERATIONS {
        let start = Instant::now();
        for (i, _el) in ids.iter().enumerate() {
            let mut i_pos = &mut pos[i];
            let mut i_vel = &mut vel[i];

            i_pos.x += i_vel.x * fixed_time;
            i_pos.y += i_vel.y * fixed_time;

            // Bump into the bounding rect
            if i_pos.x <= 0.0 || i_pos.x >= MAX_POSITION {
                i_vel.x = -i_vel.x;
            }
            if i_pos.y <= 0.0 || i_pos.y >= MAX_POSITION {
                i_vel.y = -i_vel.y;
            }
        }

        let mut death_count = 0;
        for (i, ent1) in ids.iter().enumerate() {
            let i_pos = &pos[i];
            let i_col = &col[i];
            for (j, ent2) in ids.iter().enumerate() {
                let j_pos = &pos[j];
                let j_col = &col[j];

                if ent1 == ent2 {
                    continue;
                }

                let dx = i_pos.x - j_pos.x;
                let dy = i_pos.y - j_pos.y;
                let dist_squared = (dx * dx) + (dy * dy);

                let dr = i_col.radius + j_col.radius;
                let dr_squared = dr * dr;

                if dr_squared > dist_squared {
                    cnt[i].count += 1;
                }

                if collision_limit > 0 && cnt[i].count > collision_limit {
                    death_count += 1;
                }
            }
        }

        let duration = start.elapsed();
        println!(
            "{} {:?}",
            iter_count,
            (duration.as_micros() as f64) / 1000000.0
        )
    }
}

fn native_split(size: i32, collision_limit: i32) {
    let mut ids = Vec::new();
    let mut pos_x = Vec::new();
    let mut pos_y = Vec::new();
    let mut vel_x = Vec::new();
    let mut vel_y = Vec::new();
    let mut col = Vec::new();
    let mut cnt: Vec<i32> = Vec::new();

    let mut rng = rand::thread_rng();
    for i in 0..size {
        ids.push(i + 2);
        pos_x.push(MAX_POSITION * rng.gen::<f64>());
        pos_y.push(MAX_POSITION * rng.gen::<f64>());
        vel_x.push(MAX_SPEED * rng.gen::<f64>());
        vel_y.push(MAX_SPEED * rng.gen::<f64>());
        col.push(MAX_COLLIDER * rng.gen::<f64>());
        cnt.push(0);
    }

    let fixed_time = 0.015;
    for iter_count in 0..ITERATIONS {
        let start = Instant::now();
        for (i, _el) in ids.iter().enumerate() {
            pos_x[i] += vel_x[i] * fixed_time;
            pos_y[i] += vel_y[i] * fixed_time;

            // Bump into the bounding rect
            if pos_x[i] <= 0.0 || pos_x[i] >= MAX_POSITION {
                vel_x[i] = -vel_x[i];
            }
            if pos_y[i] <= 0.0 || pos_y[i] >= MAX_POSITION {
                vel_y[i] = -vel_y[i];
            }
        }

        let mut death_count = 0;
        for (i, ent1) in ids.iter().enumerate() {
            let i_pos_x = &pos_x[i];
            let i_pos_y = &pos_y[i];
            let i_col = &col[i];

            for (j, ent2) in ids.iter().enumerate() {
                let j_pos_x = &pos_x[j];
                let j_pos_y = &pos_y[j];
                let j_col = &col[j];

                if ent1 == ent2 {
                    continue;
                }

                let dx = i_pos_x - j_pos_x;
                let dy = i_pos_y - j_pos_y;
                let dist_squared = (dx * dx) + (dy * dy);

                let dr = i_col + j_col;
                let dr_squared = dr * dr;

                if dr_squared > dist_squared {
                    cnt[i] += 1;
                }

                if collision_limit > 0 && cnt[i] > collision_limit {
                    death_count += 1;
                }
            }
        }

        let duration = start.elapsed();
        println!(
            "{} {:?}",
            iter_count,
            (duration.as_micros() as f64) / 1000000.0
        )
    }
}
