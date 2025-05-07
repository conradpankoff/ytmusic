behavior ExternalLink
  on click
    if confirm('Are you sure you want to navigate away from this page?') is false
      halt the event
    end
  end
end

behavior Video
  on keydown[(key is " " or key is "k") and target is not me] from the window
    if my paused
      halt the event
      call my play()
    else
      halt the event
      call my pause()
    end
  on keydown[key is "ArrowLeft" and not metaKey and not altKey] from the window
    halt the event
    decrement my currentTime by 5
  on keydown[key is "ArrowRight" and not metaKey and not altKey] from the window
    halt the event
    increment my currentTime by 5
  on keydown[key is "j"] from the window
    halt the event
    decrement my currentTime by 10
  on keydown[key is "l"] from the window
    halt the event
    increment my currentTime by 10
  on keydown[key is "ArrowUp" and target is not me] from the window
    halt the event
    if my volume is less than 0.9 then
      increment my volume by 0.1
    else if my volume is less than 1 then
      put 1 into my volume
    end
  on keydown[key is "ArrowDown" and target is not me] from the window
    halt the event
    if my volume is greater than 0.1
      decrement my volume by 0.1
    else if my volume is greater than 0
      put 0 into my volume
    end
  on keydown[key is "m"] from the window
    halt the event
    if my muted is true
      put false into my muted
    else
      put true into my muted
    end
  on keydown[key is "f"] from the window
    halt the event
    if document's fullscreenElement exists
      call document's exitFullscreen()
    else
      call my requestFullscreen()
    end
  end
end

behavior VideoInPlaylist
  on ended
    if #play-next exists
      call #play-next's click()
    end
  end
end

behavior Player
  init
    set :current to null
    call c_play(the first .ready)
  end

  -- playback functions

  def c_play(item)
    if item then
      if item is not :current then
        take .current from .ready
        get the first .title then put item's innerText into it
        get the first <audio/> then set its src to `/data/audio/${item's dataset's id}.mp3`
        add .current to item
        set :current to item
      end

      get the first <audio/> then if its readyState > 1 and its paused then
        call its play()
      end
    else
      get the first title then put null into it
      get the first <audio/> then set its src to null
      set :current to null
    end
  end

  def c_pause()
    get the first <audio/> then call its pause()
  end

  def c_next()
    if :current call c_play(the next .ready from :current with wrapping)
  end

  def c_prev()
    if :current call c_play(the previous .ready from :current with wrapping)
  end

  -- media events

  on canplay from <audio/>
    get the event's target
    if its paused call its play()
  end

  on ended from <audio/>
    call c_next()
  end

  -- input events

  on click from .play
    if :current call c_play(:current) else call c_play(the first .ready)
  end

  on click from .pause
    call c_pause()
  end

  on click from .back
    call c_back()
  end

  on click from .next
    call c_next()
  end

  on click from .shuffle
  end

  on click from .ready
    call c_play(the closest <li/> to the event's target)
  end
end
