use rand::prelude::*;
use std::ops::{AddAssign, Mul, Sub};
use std::time::Instant;

use crate::{ITERATIONS, MAX_COLLIDER, MAX_POSITION, MAX_SPEED};

#[derive(Copy, Clone)]
struct Point {
    x: f64,
    y: f64,
}

impl Point {
    /// Returns a new point where each value is uniformly random over the interval [0..max]
    fn rand(max: f64) -> Point {
        let mut rng = rand::thread_rng();
        Point {
            x: rng.gen::<f64>() * max,
            y: rng.gen::<f64>() * max,
        }
    }

    /// Return the square of the magnitude of the point
    fn mag_sq(&self) -> f64 {
        self.x * self.x + self.y * self.y
    }

    /// Returns the square of the distance to another point
    fn dist_sq(&self, other: &Point) -> f64 {
        (*self - *other).mag_sq()
    }
}

impl Sub for Point {
    type Output = Self;

    fn sub(self, rhs: Self) -> Self::Output {
        Point {
            x: self.x - rhs.x,
            y: self.y - rhs.y,
        }
    }
}

impl AddAssign for Point {
    fn add_assign(&mut self, rhs: Self) {
        self.x += rhs.x;
        self.y += rhs.y;
    }
}

impl Mul<f64> for Point {
    type Output = Self;

    fn mul(self, rhs: f64) -> Self::Output {
        Point {
            x: self.x * rhs,
            y: self.y * rhs,
        }
    }
}

struct Entity {
    id: usize,
    /// Counter for the number of collisions
    collisions: usize,
    /// Location of the entity
    pos: Point,
    /// Velocity of the entitiy
    vel: Point,
    radius: f64,
}

impl Entity {
    fn rand(id: usize) -> Entity {
        let mut rng = rand::thread_rng();
        Entity {
            id,
            collisions: 0,
            pos: Point::rand(MAX_POSITION),
            vel: Point::rand(MAX_SPEED),
            radius: rng.gen::<f64>() * MAX_COLLIDER,
        }
    }

    /// Apply velocity to the entity, updating position
    /// `dt` is the timestep
    fn apply_v(&mut self, dt: f64) {
        self.pos += self.vel * dt
    }

    /// Check for and process collisions with the outer bounding box, defined as 0..MAX_POSITION
    fn collide_bb(&mut self) {
        if !(0.0..MAX_POSITION).contains(&self.pos.x) {
            self.vel.x *= -1.0;
        }
        if !(0.0..MAX_POSITION).contains(&self.pos.y) {
            self.vel.y *= -1.0;
        }
    }

    /// Returns the square of the distance between two entities
    fn dist_sq(&self, other: &Entity) -> f64 {
        self.pos.dist_sq(&other.pos)
    }

    /// Update position, applying the velocity and then performing boundary checks
    /// `dt` is the timestamp
    fn update_pos(&mut self, dt: f64) {
        self.apply_v(dt);
        self.collide_bb();
    }

    /// Tests if two entities are colliding
    fn colliding(&self, other: &Entity) -> bool {
        let dr = (self.radius + other.radius).powi(2);
        self.dist_sq(other) > dr
    }

    fn inc_collision(&mut self, collision_limit: usize, death_count: &mut usize) {
        self.collisions += 1;
        if self.collisions > collision_limit {
            // Removed because the reference impl does not have this, but it seems like
            // it make sense
            // self.collision = 0;
            *death_count += 1;
        }
    }

    /// Process a possible collision between two entities
    /// If the two entities are not in contact, nothing changes
    fn collide(&mut self, other: &mut Entity, collision_limit: usize, death_count: &mut usize) {
        if self.colliding(other) {
            self.inc_collision(collision_limit, death_count);
            other.inc_collision(collision_limit, death_count);
        }
    }
}

pub fn native(size: usize, collision_limit: usize) {
    let mut entities = Vec::with_capacity(size);

    for i in 0..size {
        entities.push(Entity::rand(i + 2));
    }

    let fixed_dt = 0.015;
    let mut death_count = 0;

    for step in 0..ITERATIONS {
        let start = Instant::now();

        entities
            .iter_mut()
            .for_each(|entity| entity.update_pos(fixed_dt));

        for i in 0..size - 1 {
            let (left, right) = entities.split_at_mut(i + 1);
            let e0 = &mut left[i];
            for e1 in right {
                e0.collide(e1, collision_limit, &mut death_count);
            }
        }

        let elapsed = start.elapsed();
        println!("{step}: {:?}", (elapsed.as_micros() as f64) / 1000000.0);
    }
}
