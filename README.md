# doors

## Description

## What you need

* [PhidgetSBC3](http://www.phidgets.com/products.php?product_id=1073_0)
* Relays (i.e. [Dual SSR Relay Board](http://www.phidgets.com/products.php?category=9&product_id=3053_0) or [8 Channel SSR Module Board](http://www.sainsmart.com/8-channel-5v-solid-state-relay-module-board-omron-ssr-4-pic-arm-avr-dsp-arduino.html))
* Door Controller (i.e [KT-300](http://www.kantech.com/Products/controllers_kt300.aspx))

## Preperation

### Hardware

This will be mostly up to you because it will vary widly depending on what door controller you use.
What you need to do is connect the PhidgetSBC3 to the relays and the relays to the door controller in such a way that when closing the relay, the door will be unlocked.
I recommend the [Kantech KT-300](http://www.kantech.com/Products/controllers_kt300.aspx) and [Kantech EntraPass](http://www.kantech.com/Products/software_entrapass.aspx) because that is what we use at Shopify and we've got it working. The KT-300 has auxiliary inputs that can be configured as secondary Request to exit (REX) for each door.

### Software

1. Plug your PhidgetSBC3 into power and network
2. Go to [phidgetsbc.local](http://phidgetsbc.local/) and set a password
<a href="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.53.39_AM.png"><img src="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.53.39_AM.png" alt="Go to phidgetsbc.local and set a password" width="828" height="363" class="size-full wp-image-768" /></a>
3. Enable the SSH server
<a href="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.54.05_AM.png"><img src="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.54.05_AM.png" alt="Enable the SSH server" width="828" height="656" class="size-full wp-image-769" /></a>
4. Enable the full Debian Package Repository
<a href="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.58.28_AM.png"><img src="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.58.28_AM.png" alt="Enable the full Debian Package Repository " width="828" height="748" class="size-full wp-image-772" /></a>
5. Add your ssh key to the phidgetsbc *(optional)*
<pre><code>ssh root@phidgetsbc.local "mkdir -p .ssh && echo '$(cat ~/.ssh/id_rsa.pub)' >> .ssh/authorized_keys"</code></pre>
6. SSH in
<pre><code>ssh root@phidgetsbc.local</code></pre>
7. Add my apt key
<pre><code>apt-key adv --keyserver keys.gnupg.net --recv-keys B4F808A2</code></pre>
8. Add my apt repo
<pre><code>echo "deb http://apt.samuelkadolph.com/ wheezy main" > /etc/apt/sources.list.d/samuelkadolph.list</code></pre>
9. Update apt
<pre><code>apt-get update</code></pre>
10. Install the required packages
<pre><code>apt-get install golang-tip build-essential git-core libphidget21-dev ca-certificates -y</code></pre>

## Installation

1. Go to [phidgetsbc.local](http://phidgetsbc.local/) and log in
<a href="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.35.27_AM.png"><img src="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.35.27_AM.png" alt="Go to phidgetsbc.local and log in" width="828" height="299" /></a>
2. Create a project
<a href="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.38.37_AM.png"><img src="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.38.37_AM.png" alt="Create a project" width="828" height="517" /></a>
3. Install the application
<pre><code>GOPATH=$HOME/go go get github.com/samuelkadolph/doors</code></pre>
4. Copy the binary to the project
<pre><code>cp $HOME/go/bin/doors /usr/userapps/doors/doors</code></pre>
5. Create the config file
<pre><code>TODO</code></pre>
6. Set the executable and enable boot startup for the project
<a href="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.39.50_AM.png"><img src="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.39.50_AM.png" alt="Set the executable and enable boot startup for the project" width="828" height="787" /></a>
7. Start the application
<a href="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.40.49_AM.png"><img src="https://i0.wp.com/samuel.kadolph.com/wp-content/uploads/2013/03/Screen_Shot_2013-03-25_at_1.40.49_AM.png" alt="Start the application" width="828" height="506" /></a>
8. Open a door!
<pre><code>curl http://phidgetsbc.local:4567/doors/XXX/unlock -d "secret=SECRET"</code></pre>

## API

Secret should be passed in as a query string or as a url encoded body. All responses are in JSON.

#### Actions

    GET /doors
> Gets all the doors.<br />
> Returns: array of hashes

    GET /doors/{door}
> Gets a door.<br />
> Returns: hash with `id`, `lock`, `mag`, and `name` fields

    POST /doors/{door}/unlock
> Unlocks a door.<br />
> Returns: hash with `success` field and possibly `error` field

    POST /doors/{door}/mag/engage
> Engages the mag for a door.<br />
> Returns: hash with `success` field and possibly `error` field

    POST /doors/{door}/mag/disengage
> Disengages the mag for a door.<br />
> Returns: hash with `success` field and possibly `error` field

#### Fields

    error
> Error message.<br />
> Type: String

    id
> ID of a door.<br />
> Type: String

    lock
> Status of the lock of a door.<br />
> Type: String<br />
> Values: `error`, `locked`, `unlocked`, `unsupported`

    mag
> Status of a mag of a door.<br />
> Type: String<br />
> Values: `disengaged`, `engaged`, `error`, `unsupported`

    name
> Nice name of a door.<br />
> Type: String

    success
> Result of an action.<br />
> Type: Boolean
